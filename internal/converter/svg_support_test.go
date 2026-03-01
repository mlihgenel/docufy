package converter

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestImageConverterSupportsSVGSources(t *testing.T) {
	ic := &ImageConverter{}

	if !ic.SupportsConversion("svg", "png") {
		t.Fatalf("expected svg -> png to be supported")
	}
	if !ic.SupportsConversion("svg", "pdf") {
		t.Fatalf("expected svg -> pdf to be supported")
	}
	if ic.SupportsConversion("svg", "svg") {
		t.Fatalf("did not expect svg -> svg to be supported")
	}
}

func TestIsResizableFormatSupportsSVG(t *testing.T) {
	if !IsResizableFormat("svg") {
		t.Fatalf("expected svg to be resizable")
	}
}

func TestSVGToPNGConversion(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "shape.svg")
	outputPath := filepath.Join(dir, "shape.png")

	if err := os.WriteFile(inputPath, []byte(simpleRectSVG(48, 24)), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ic := &ImageConverter{}
	if err := ic.Convert(inputPath, outputPath, Options{}); err != nil {
		t.Fatalf("svg -> png failed: %v", err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode config failed: %v", err)
	}
	if cfg.Width != 48 || cfg.Height != 24 {
		t.Fatalf("expected 48x24 output, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestSVGResizeToPNG(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "shape.svg")
	outputPath := filepath.Join(dir, "resized.png")

	if err := os.WriteFile(inputPath, []byte(simpleRectSVG(48, 24)), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ic := &ImageConverter{}
	if err := ic.Convert(inputPath, outputPath, Options{
		Resize: &ResizeSpec{Width: 120, Height: 120, Mode: ResizeModePad},
	}); err != nil {
		t.Fatalf("svg resize failed: %v", err)
	}

	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	defer f.Close()

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode config failed: %v", err)
	}
	if cfg.Width != 120 || cfg.Height != 120 {
		t.Fatalf("expected 120x120 output, got %dx%d", cfg.Width, cfg.Height)
	}
}

func TestSVGToPDFConversion(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "shape.svg")
	outputPath := filepath.Join(dir, "shape.pdf")

	if err := os.WriteFile(inputPath, []byte(simpleRectSVG(64, 64)), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	ic := &ImageConverter{}
	if err := ic.Convert(inputPath, outputPath, Options{}); err != nil {
		t.Fatalf("svg -> pdf failed: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty pdf output")
	}
}

func simpleRectSVG(width int, height int) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<svg width="` + itoa(width) + `" height="` + itoa(height) + `" xmlns="http://www.w3.org/2000/svg">` +
		`<rect x="0" y="0" width="` + itoa(width) + `" height="` + itoa(height) + `"/>` +
		`</svg>`
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
