package converter

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/image/vector"
)

const (
	svgDefaultDPI      = 96.0
	svgBezierArcFactor = 0.5522847498307936
)

type svgGraphic struct {
	Width   float64
	Height  float64
	OffsetX float64
	OffsetY float64
	Paths   [][]gofpdf.SVGBasicSegmentType
}

func decodeSVGFile(path string) (image.Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("svg okunamadı: %w", err)
	}

	graphic, err := parseSVGGraphic(data)
	if err != nil {
		return nil, err
	}

	return rasterizeSVGGraphic(graphic, 0, 0)
}

func svgDimensionsFromFile(path string) (int, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, fmt.Errorf("svg okunamadı: %w", err)
	}

	graphic, err := parseSVGGraphic(data)
	if err != nil {
		return 0, 0, err
	}

	return maxInt(1, int(math.Round(graphic.Width))), maxInt(1, int(math.Round(graphic.Height))), nil
}

func encodeImagePDF(path string, img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return fmt.Errorf("pdf için geçersiz görsel boyutu")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("görsel pdf için hazırlanamadı: %w", err)
	}

	pageW := pixelsToMillimeters(bounds.Dx())
	pageH := pixelsToMillimeters(bounds.Dy())

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: orientationForSize(pageW, pageH),
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: pageW, Ht: pageH},
	})
	pdf.SetMargins(0, 0, 0)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	opts := gofpdf.ImageOptions{ImageType: "png", ReadDpi: false}
	pdf.RegisterImageOptionsReader("svg-raster", opts, bytes.NewReader(buf.Bytes()))
	pdf.ImageOptions("svg-raster", 0, 0, pageW, pageH, false, opts, 0, "")

	if err := pdf.Error(); err != nil {
		return fmt.Errorf("pdf oluşturulamadı: %w", err)
	}
	return pdf.OutputFileAndClose(path)
}

func parseSVGGraphic(data []byte) (svgGraphic, error) {
	var graphic svgGraphic
	var viewBoxSet bool
	var viewBoxWidth, viewBoxHeight float64

	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return svgGraphic{}, fmt.Errorf("svg parse hatası: %w", err)
		}

		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}

		name := strings.ToLower(start.Name.Local)
		attrs := xmlAttrMap(start.Attr)

		if name == "svg" {
			if w, h, offX, offY, ok := parseSVGRootDimensions(attrs); ok {
				graphic.Width = w
				graphic.Height = h
				graphic.OffsetX = offX
				graphic.OffsetY = offY
			}
			if viewBox, ok := attrs["viewbox"]; ok {
				if minX, minY, w, h, ok := parseSVGViewBox(viewBox); ok {
					viewBoxSet = true
					viewBoxWidth = w
					viewBoxHeight = h
					if graphic.OffsetX == 0 && graphic.OffsetY == 0 {
						graphic.OffsetX = minX
						graphic.OffsetY = minY
					}
				}
			}
			continue
		}

		pathData, ok := svgElementPathData(name, attrs)
		if !ok {
			continue
		}

		segments, err := parseSVGPathData(pathData)
		if err != nil {
			return svgGraphic{}, err
		}
		if len(segments) > 0 {
			graphic.Paths = append(graphic.Paths, segments)
		}
	}

	if graphic.Width <= 0 || graphic.Height <= 0 {
		if viewBoxSet && viewBoxWidth > 0 && viewBoxHeight > 0 {
			graphic.Width = viewBoxWidth
			graphic.Height = viewBoxHeight
		}
	}
	if graphic.Width <= 0 || graphic.Height <= 0 {
		return svgGraphic{}, fmt.Errorf("svg boyutları çözümlenemedi")
	}
	if len(graphic.Paths) == 0 {
		return svgGraphic{}, fmt.Errorf("svg içinde desteklenen çizim elemanı bulunamadı")
	}

	return graphic, nil
}

func rasterizeSVGGraphic(graphic svgGraphic, targetWidth int, targetHeight int) (image.Image, error) {
	width := targetWidth
	height := targetHeight
	if width <= 0 {
		width = maxInt(1, int(math.Round(graphic.Width)))
	}
	if height <= 0 {
		height = maxInt(1, int(math.Round(graphic.Height)))
	}

	canvas := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float32(float64(width) / graphic.Width)
	scaleY := float32(float64(height) / graphic.Height)
	fill := image.NewUniform(color.Black)

	xVal := func(x float64) float32 {
		return float32((x - graphic.OffsetX)) * scaleX
	}
	yVal := func(y float64) float32 {
		return float32((y - graphic.OffsetY)) * scaleY
	}

	for _, path := range graphic.Paths {
		if len(path) == 0 {
			continue
		}

		rasterizer := vector.NewRasterizer(width, height)
		for _, seg := range path {
			switch seg.Cmd {
			case 'M':
				rasterizer.MoveTo(xVal(seg.Arg[0]), yVal(seg.Arg[1]))
			case 'L':
				rasterizer.LineTo(xVal(seg.Arg[0]), yVal(seg.Arg[1]))
			case 'C':
				rasterizer.CubeTo(
					xVal(seg.Arg[0]), yVal(seg.Arg[1]),
					xVal(seg.Arg[2]), yVal(seg.Arg[3]),
					xVal(seg.Arg[4]), yVal(seg.Arg[5]),
				)
			case 'Q':
				rasterizer.QuadTo(
					xVal(seg.Arg[0]), yVal(seg.Arg[1]),
					xVal(seg.Arg[2]), yVal(seg.Arg[3]),
				)
			case 'H':
				_, currentY := rasterizer.Pen()
				rasterizer.LineTo(xVal(seg.Arg[0]), currentY)
			case 'V':
				currentX, _ := rasterizer.Pen()
				rasterizer.LineTo(currentX, yVal(seg.Arg[0]))
			case 'Z':
				rasterizer.ClosePath()
			}
		}
		rasterizer.Draw(canvas, canvas.Bounds(), fill, image.Point{})
	}

	return canvas, nil
}

func parseSVGRootDimensions(attrs map[string]string) (width float64, height float64, offsetX float64, offsetY float64, ok bool) {
	if rawViewBox, hasViewBox := attrs["viewbox"]; hasViewBox {
		if minX, minY, vbWidth, vbHeight, parsed := parseSVGViewBox(rawViewBox); parsed {
			offsetX = minX
			offsetY = minY
			if width == 0 {
				width = vbWidth
			}
			if height == 0 {
				height = vbHeight
			}
		}
	}

	if rawWidth, exists := attrs["width"]; exists {
		if parsedWidth, parsed := parseSVGLength(rawWidth); parsed {
			width = parsedWidth
		}
	}
	if rawHeight, exists := attrs["height"]; exists {
		if parsedHeight, parsed := parseSVGLength(rawHeight); parsed {
			height = parsedHeight
		}
	}

	if width > 0 && height > 0 {
		return width, height, offsetX, offsetY, true
	}
	return 0, 0, offsetX, offsetY, false
}

func parseSVGViewBox(raw string) (minX float64, minY float64, width float64, height float64, ok bool) {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	if len(parts) != 4 {
		return 0, 0, 0, 0, false
	}

	values := make([]float64, 4)
	for i, part := range parts {
		v, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return 0, 0, 0, 0, false
		}
		values[i] = v
	}

	if values[2] <= 0 || values[3] <= 0 {
		return 0, 0, 0, 0, false
	}
	return values[0], values[1], values[2], values[3], true
}

func parseSVGLength(raw string) (float64, bool) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" || strings.HasSuffix(value, "%") {
		return 0, false
	}

	units := []struct {
		suffix string
		scale  float64
	}{
		{suffix: "px", scale: 1},
		{suffix: "pt", scale: svgDefaultDPI / 72.0},
		{suffix: "pc", scale: svgDefaultDPI / 6.0},
		{suffix: "mm", scale: svgDefaultDPI / 25.4},
		{suffix: "cm", scale: svgDefaultDPI / 2.54},
		{suffix: "in", scale: svgDefaultDPI},
	}

	for _, unit := range units {
		if strings.HasSuffix(value, unit.suffix) {
			number, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(value, unit.suffix)), 64)
			if err != nil || number <= 0 {
				return 0, false
			}
			return number * unit.scale, true
		}
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil || number <= 0 {
		return 0, false
	}
	return number, true
}

func xmlAttrMap(attrs []xml.Attr) map[string]string {
	result := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		result[strings.ToLower(attr.Name.Local)] = attr.Value
	}
	return result
}

func svgElementPathData(name string, attrs map[string]string) (string, bool) {
	switch name {
	case "path":
		d := strings.TrimSpace(attrs["d"])
		return d, d != ""
	case "rect":
		return rectToSVGPath(attrs)
	case "line":
		return lineToSVGPath(attrs)
	case "polyline":
		return polyPointsToSVGPath(attrs["points"], false)
	case "polygon":
		return polyPointsToSVGPath(attrs["points"], true)
	case "circle":
		return circleToSVGPath(attrs)
	case "ellipse":
		return ellipseToSVGPath(attrs)
	default:
		return "", false
	}
}

func parseSVGPathData(pathData string) ([]gofpdf.SVGBasicSegmentType, error) {
	escaped := xmlEscapeAttr(pathData)
	buf := []byte(fmt.Sprintf(`<svg width="1" height="1"><path d="%s"/></svg>`, escaped))
	parsed, err := gofpdf.SVGBasicParse(buf)
	if err != nil {
		return nil, fmt.Errorf("svg path parse hatası: %w", err)
	}
	if len(parsed.Segments) == 0 {
		return nil, fmt.Errorf("svg path boş")
	}
	return parsed.Segments[0], nil
}

func rectToSVGPath(attrs map[string]string) (string, bool) {
	x, _ := parseOptionalSVGNumber(attrs["x"])
	y, _ := parseOptionalSVGNumber(attrs["y"])
	width, okWidth := parseRequiredSVGNumber(attrs["width"])
	height, okHeight := parseRequiredSVGNumber(attrs["height"])
	if !okWidth || !okHeight {
		return "", false
	}

	return fmt.Sprintf(
		"M %g %g L %g %g L %g %g L %g %g Z",
		x, y,
		x+width, y,
		x+width, y+height,
		x, y+height,
	), true
}

func lineToSVGPath(attrs map[string]string) (string, bool) {
	x1, okX1 := parseRequiredSVGNumber(attrs["x1"])
	y1, okY1 := parseRequiredSVGNumber(attrs["y1"])
	x2, okX2 := parseRequiredSVGNumber(attrs["x2"])
	y2, okY2 := parseRequiredSVGNumber(attrs["y2"])
	if !okX1 || !okY1 || !okX2 || !okY2 {
		return "", false
	}

	return fmt.Sprintf("M %g %g L %g %g", x1, y1, x2, y2), true
}

func polyPointsToSVGPath(raw string, closePath bool) (string, bool) {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n'
	})
	if len(parts) < 4 || len(parts)%2 != 0 {
		return "", false
	}

	var b strings.Builder
	for i := 0; i < len(parts); i += 2 {
		x, errX := strconv.ParseFloat(parts[i], 64)
		y, errY := strconv.ParseFloat(parts[i+1], 64)
		if errX != nil || errY != nil {
			return "", false
		}
		if i == 0 {
			fmt.Fprintf(&b, "M %g %g", x, y)
			continue
		}
		fmt.Fprintf(&b, " L %g %g", x, y)
	}
	if closePath {
		b.WriteString(" Z")
	}
	return b.String(), true
}

func circleToSVGPath(attrs map[string]string) (string, bool) {
	cx, okCX := parseRequiredSVGNumber(attrs["cx"])
	cy, okCY := parseRequiredSVGNumber(attrs["cy"])
	r, okR := parseRequiredSVGNumber(attrs["r"])
	if !okCX || !okCY || !okR || r <= 0 {
		return "", false
	}
	return ellipsePath(cx, cy, r, r), true
}

func ellipseToSVGPath(attrs map[string]string) (string, bool) {
	cx, okCX := parseRequiredSVGNumber(attrs["cx"])
	cy, okCY := parseRequiredSVGNumber(attrs["cy"])
	rx, okRX := parseRequiredSVGNumber(attrs["rx"])
	ry, okRY := parseRequiredSVGNumber(attrs["ry"])
	if !okCX || !okCY || !okRX || !okRY || rx <= 0 || ry <= 0 {
		return "", false
	}
	return ellipsePath(cx, cy, rx, ry), true
}

func ellipsePath(cx float64, cy float64, rx float64, ry float64) string {
	kx := rx * svgBezierArcFactor
	ky := ry * svgBezierArcFactor

	return fmt.Sprintf(
		"M %g %g "+
			"C %g %g %g %g %g %g "+
			"C %g %g %g %g %g %g "+
			"C %g %g %g %g %g %g "+
			"C %g %g %g %g %g %g Z",
		cx+rx, cy,
		cx+rx, cy+ky, cx+kx, cy+ry, cx, cy+ry,
		cx-kx, cy+ry, cx-rx, cy+ky, cx-rx, cy,
		cx-rx, cy-ky, cx-kx, cy-ry, cx, cy-ry,
		cx+kx, cy-ry, cx+rx, cy-ky, cx+rx, cy,
	)
}

func parseRequiredSVGNumber(raw string) (float64, bool) {
	return parseSVGLength(raw)
}

func parseOptionalSVGNumber(raw string) (float64, bool) {
	if strings.TrimSpace(raw) == "" {
		return 0, true
	}
	return parseSVGLength(raw)
}

func xmlEscapeAttr(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		`"`, "&quot;",
		"'", "&apos;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

func detectSVGByContent(header []byte) string {
	trimmed := strings.TrimSpace(strings.TrimPrefix(string(header), "\uFEFF"))
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "<svg") {
		return "svg"
	}
	if strings.HasPrefix(lower, "<?xml") {
		if idx := strings.Index(lower, "<svg"); idx >= 0 && idx < 512 {
			return "svg"
		}
	}
	return ""
}

func pixelsToMillimeters(px int) float64 {
	return float64(px) * 25.4 / svgDefaultDPI
}

func orientationForSize(width float64, height float64) string {
	if width > height {
		return "L"
	}
	return "P"
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
