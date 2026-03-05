package ai

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mlihgenel/docufy/v2/internal/converter"
)

const (
	ToolConvertFile  = "convert_file"
	ToolTrimVideo    = "trim_video"
	ToolExtractAudio = "extract_audio"
	ToolGetFileInfo  = "get_file_info"
)

type Gateway struct {
	policy Policy
}

func NewGateway(policy Policy) *Gateway {
	return &Gateway{policy: policy}
}

func (g *Gateway) ValidatePath(path string) (string, error) {
	return g.policy.ValidatePath(path)
}

func (g *Gateway) ConvertFile(req ConvertFileRequest) (ToolResult, error) {
	inputPath, err := g.policy.ValidatePath(req.InputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, Error: err.Error()}, err
	}
	if _, err := os.Stat(inputPath); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}

	to := converter.NormalizeFormat(req.To)
	if to == "" {
		err := fmt.Errorf("to zorunlu")
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}

	outputPath, err := g.resolveOutputPath(inputPath, to, req.OutputPath, req.OutputDir, req.Name)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}

	policy := converter.NormalizeConflictPolicy(req.OnConflict)
	resolvedOutput, skip, err := converter.ResolveOutputPathConflict(outputPath, policy)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, OutputPath: outputPath, Error: err.Error()}, err
	}
	if skip {
		return ToolResult{
			Status:     StatusSkipped,
			Tool:       ToolConvertFile,
			InputPath:  inputPath,
			OutputPath: outputPath,
			Message:    "output exists (policy: skip)",
		}, nil
	}

	from := converter.DetectFormat(inputPath)
	if from == "" {
		err := fmt.Errorf("kaynak format algilanamadi")
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}
	conv, err := converter.FindConverter(from, to)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}

	metadataMode := converter.NormalizeMetadataMode(req.MetadataMode)
	if metadataMode == "" {
		err := fmt.Errorf("gecersiz metadata_mode: %s", req.MetadataMode)
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, Error: err.Error()}, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedOutput), 0755); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	opts := converter.Options{
		Quality:      req.Quality,
		Verbose:      false,
		Name:         req.Name,
		MetadataMode: metadataMode,
	}
	if err := conv.Convert(inputPath, resolvedOutput, opts); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolConvertFile, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	return ToolResult{
		Status:     StatusSuccess,
		Tool:       ToolConvertFile,
		InputPath:  inputPath,
		OutputPath: resolvedOutput,
		Message:    "conversion completed",
	}, nil
}

func (g *Gateway) GetFileInfo(req FileInfoRequest) (ToolResult, error) {
	path, err := g.policy.ValidatePath(req.Path)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolGetFileInfo, Error: err.Error()}, err
	}
	info, err := converter.GetFileInfo(path)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolGetFileInfo, InputPath: path, Error: err.Error()}, err
	}
	return ToolResult{
		Status:    StatusSuccess,
		Tool:      ToolGetFileInfo,
		InputPath: path,
		FileInfo:  &info,
	}, nil
}

func (g *Gateway) TrimVideo(req TrimVideoRequest) (ToolResult, error) {
	inputPath, err := g.policy.ValidatePath(req.InputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, Error: err.Error()}, err
	}
	if _, err := os.Stat(inputPath); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "clip"
	}
	if mode != "clip" && mode != "remove" {
		err := fmt.Errorf("gecersiz mode: %s", req.Mode)
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}

	if strings.TrimSpace(req.End) != "" && strings.TrimSpace(req.Duration) != "" {
		err := fmt.Errorf("end ve duration birlikte kullanilamaz")
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}

	start := strings.TrimSpace(req.Start)
	if start == "" {
		start = "0"
	}
	if mode == "remove" && strings.TrimSpace(req.End) == "" && strings.TrimSpace(req.Duration) == "" {
		err := fmt.Errorf("remove mode icin end veya duration gerekli")
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}
	if mode == "clip" && strings.TrimSpace(req.End) == "" && strings.TrimSpace(req.Duration) == "" {
		req.Duration = "10"
	}

	targetFormat := converter.NormalizeFormat(req.To)
	if targetFormat == "" {
		targetFormat = converter.NormalizeFormat(converter.DetectFormat(inputPath))
	}
	if targetFormat == "" {
		err := fmt.Errorf("hedef format belirlenemedi")
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}

	outputPath, err := g.resolveOutputPath(inputPath, targetFormat, req.OutputPath, req.OutputDir, req.Name)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}
	conflict := converter.NormalizeConflictPolicy(req.OnConflict)
	resolvedOutput, skip, err := converter.ResolveOutputPathConflict(outputPath, conflict)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, OutputPath: outputPath, Error: err.Error()}, err
	}
	if skip {
		return ToolResult{
			Status:     StatusSkipped,
			Tool:       ToolTrimVideo,
			InputPath:  inputPath,
			OutputPath: outputPath,
			Message:    "output exists (policy: skip)",
		}, nil
	}

	metadataMode := converter.NormalizeMetadataMode(req.MetadataMode)
	if metadataMode == "" {
		err := fmt.Errorf("gecersiz metadata_mode: %s", req.MetadataMode)
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}

	codec := normalizeTrimCodec(req.Codec)
	if codec == "" {
		err := fmt.Errorf("gecersiz codec modu: %s", req.Codec)
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: err.Error()}, err
	}
	if mode == "remove" && codec == "copy" {
		// Remove mode filter graph gerektirir; copy codec bu durumda gecersizdir.
		codec = "reencode"
	}

	if err := os.MkdirAll(filepath.Dir(resolvedOutput), 0755); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, Error: "ffmpeg bulunamadi"}, err
	}

	if mode == "remove" {
		if err := runTrimRemoveFFmpeg(
			ffmpegPath,
			inputPath,
			resolvedOutput,
			start,
			strings.TrimSpace(req.End),
			strings.TrimSpace(req.Duration),
			targetFormat,
			codec,
			req.Quality,
			metadataMode,
		); err != nil {
			return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
		}
		return ToolResult{
			Status:     StatusSuccess,
			Tool:       ToolTrimVideo,
			InputPath:  inputPath,
			OutputPath: resolvedOutput,
			Message:    "video trim remove completed",
		}, nil
	}

	args := []string{"-loglevel", "error", "-i", inputPath, "-ss", start}
	if strings.TrimSpace(req.End) != "" {
		args = append(args, "-to", strings.TrimSpace(req.End))
	}
	if strings.TrimSpace(req.Duration) != "" {
		args = append(args, "-t", strings.TrimSpace(req.Duration))
	}
	if codec == "copy" {
		args = append(args, "-c", "copy")
	} else {
		args = append(args, trimReencodeArgs(req.Quality, targetFormat != "gif")...)
	}
	args = append(args, converter.MetadataFFmpegArgs(metadataMode)...)
	args = append(args, "-y", resolvedOutput)

	if out, err := exec.Command(ffmpegPath, args...).CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return ToolResult{Status: StatusFailed, Tool: ToolTrimVideo, InputPath: inputPath, OutputPath: resolvedOutput, Error: msg}, err
	}

	return ToolResult{
		Status:     StatusSuccess,
		Tool:       ToolTrimVideo,
		InputPath:  inputPath,
		OutputPath: resolvedOutput,
		Message:    "video trim completed",
	}, nil
}

func (g *Gateway) ExtractAudio(req ExtractAudioRequest) (ToolResult, error) {
	inputPath, err := g.policy.ValidatePath(req.InputPath)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, Error: err.Error()}, err
	}
	if _, err := os.Stat(inputPath); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, Error: err.Error()}, err
	}

	targetFormat := converter.NormalizeFormat(req.To)
	if targetFormat == "" {
		targetFormat = "mp3"
	}
	if !isValidExtractAudioFormat(targetFormat) {
		err := fmt.Errorf("desteklenmeyen ses formati: %s", targetFormat)
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, Error: err.Error()}, err
	}

	outputPath, err := g.resolveOutputPath(inputPath, targetFormat, req.OutputPath, req.OutputDir, req.Name)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, Error: err.Error()}, err
	}
	conflict := converter.NormalizeConflictPolicy(req.OnConflict)
	resolvedOutput, skip, err := converter.ResolveOutputPathConflict(outputPath, conflict)
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, OutputPath: outputPath, Error: err.Error()}, err
	}
	if skip {
		return ToolResult{
			Status:     StatusSkipped,
			Tool:       ToolExtractAudio,
			InputPath:  inputPath,
			OutputPath: outputPath,
			Message:    "output exists (policy: skip)",
		}, nil
	}

	metadataMode := converter.NormalizeMetadataMode(req.MetadataMode)
	if metadataMode == "" {
		err := fmt.Errorf("gecersiz metadata_mode: %s", req.MetadataMode)
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, Error: err.Error()}, err
	}

	if err := os.MkdirAll(filepath.Dir(resolvedOutput), 0755); err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, OutputPath: resolvedOutput, Error: err.Error()}, err
	}

	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, Error: "ffmpeg bulunamadi"}, err
	}

	args := []string{"-loglevel", "error", "-i", inputPath, "-vn"}
	args = append(args, extractAudioCodecArgs(targetFormat, req.Quality, req.Copy)...)
	args = append(args, converter.MetadataFFmpegArgs(metadataMode)...)
	args = append(args, "-y", resolvedOutput)

	if out, err := exec.Command(ffmpegPath, args...).CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return ToolResult{Status: StatusFailed, Tool: ToolExtractAudio, InputPath: inputPath, OutputPath: resolvedOutput, Error: msg}, err
	}

	return ToolResult{
		Status:     StatusSuccess,
		Tool:       ToolExtractAudio,
		InputPath:  inputPath,
		OutputPath: resolvedOutput,
		Message:    "audio extraction completed",
	}, nil
}

func (g *Gateway) resolveOutputPath(inputPath string, to string, outputPath string, outputDir string, name string) (string, error) {
	if strings.TrimSpace(outputPath) == "" {
		outputPath = converter.BuildOutputPath(inputPath, outputDir, to, name)
	}
	return g.policy.ValidatePath(outputPath)
}

func normalizeTrimCodec(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "", "auto", "copy":
		return "copy"
	case "reencode":
		return "reencode"
	default:
		return ""
	}
}

func videoCRF(quality int) int {
	if quality <= 0 {
		return 23
	}
	switch {
	case quality <= 25:
		return 30
	case quality <= 50:
		return 27
	case quality <= 75:
		return 24
	default:
		return 20
	}
}

func trimReencodeArgs(quality int, includeAudio bool) []string {
	args := []string{
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", strconv.Itoa(videoCRF(quality)),
	}
	if includeAudio {
		args = append(args, "-c:a", "aac", "-b:a", "128k")
	} else {
		args = append(args, "-an")
	}
	return args
}

func runTrimRemoveFFmpeg(ffmpegPath string, inputPath string, outputPath string, start string, end string, duration string, targetFormat string, codec string, quality int, metadataMode string) error {
	startSec, err := parseTrimTimeToSeconds(start)
	if err != nil {
		return fmt.Errorf("gecersiz baslangic zamani: %s", start)
	}
	endSec := 0.0
	if strings.TrimSpace(end) != "" {
		endSec, err = parseTrimTimeToSeconds(end)
		if err != nil {
			return fmt.Errorf("gecersiz bitis zamani: %s", end)
		}
	} else if strings.TrimSpace(duration) != "" {
		durationSec, parseErr := parseTrimTimeToSeconds(duration)
		if parseErr != nil {
			return fmt.Errorf("gecersiz sure degeri: %s", duration)
		}
		endSec = startSec + durationSec
	} else {
		return fmt.Errorf("remove mode icin end veya duration gerekli")
	}
	if endSec <= startSec {
		return fmt.Errorf("bitis zamani baslangictan buyuk olmali")
	}

	if sourceDuration, ok := probeMediaDurationSeconds(inputPath); ok {
		const epsilon = 0.001
		if startSec >= sourceDuration-epsilon {
			return fmt.Errorf("baslangic zamani video suresini asiyor")
		}
		if endSec > sourceDuration {
			endSec = sourceDuration
		}
		if endSec <= startSec+epsilon {
			return fmt.Errorf("bitis zamani baslangictan buyuk olmali")
		}
		if startSec <= epsilon && endSec >= sourceDuration-epsilon {
			return fmt.Errorf("silinecek aralik tum videoyu kapsiyor")
		}
	}

	if codec == "copy" {
		return fmt.Errorf("remove mode copy codec ile calismaz")
	}

	startMark := formatSecondsForFFmpeg(startSec)
	endMark := formatSecondsForFFmpeg(endSec)
	videoFilter := fmt.Sprintf("select='not(between(t,%s,%s))',setpts=N/FRAME_RATE/TB", startMark, endMark)
	audioFilter := fmt.Sprintf("aselect='not(between(t,%s,%s))',asetpts=N/SR/TB", startMark, endMark)

	hasAudio, hasAudioKnown := probeHasAudioStream(inputPath)
	includeAudio := hasAudioKnown && hasAudio && converter.NormalizeFormat(targetFormat) != "gif"

	args := []string{"-loglevel", "error", "-i", inputPath, "-vf", videoFilter}
	if includeAudio {
		args = append(args, "-af", audioFilter)
	}
	args = append(args, trimReencodeArgs(quality, includeAudio)...)
	args = append(args, converter.MetadataFFmpegArgs(metadataMode)...)
	args = append(args, "-y", outputPath)

	if out, runErr := exec.Command(ffmpegPath, args...).CombinedOutput(); runErr != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = runErr.Error()
		}
		return fmt.Errorf("%s", msg)
	}

	return nil
}

func parseTrimTimeToSeconds(value string) (float64, error) {
	normalized := strings.TrimSpace(strings.ReplaceAll(value, ",", "."))
	if normalized == "" {
		return 0, fmt.Errorf("bos zaman")
	}

	if strings.Contains(normalized, ":") {
		parts := strings.Split(normalized, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return 0, fmt.Errorf("zaman formati hatali")
		}

		values := make([]float64, 0, len(parts))
		for _, part := range parts {
			n := strings.TrimSpace(part)
			if n == "" {
				return 0, fmt.Errorf("zaman formati hatali")
			}
			v, err := strconv.ParseFloat(n, 64)
			if err != nil || v < 0 {
				return 0, fmt.Errorf("zaman formati hatali")
			}
			values = append(values, v)
		}

		if len(values) == 2 {
			if values[1] >= 60 {
				return 0, fmt.Errorf("saniye 60'tan kucuk olmali")
			}
			return values[0]*60 + values[1], nil
		}

		if values[1] >= 60 || values[2] >= 60 {
			return 0, fmt.Errorf("dakika/saniye 60'tan kucuk olmali")
		}
		return values[0]*3600 + values[1]*60 + values[2], nil
	}

	seconds, err := strconv.ParseFloat(normalized, 64)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("gecersiz sayi")
	}
	return seconds, nil
}

func formatSecondsForFFmpeg(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func probeMediaDurationSeconds(inputPath string) (float64, bool) {
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return 0, false
	}
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, false
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || seconds <= 0 {
		return 0, false
	}
	return seconds, true
}

func probeHasAudioStream(inputPath string) (bool, bool) {
	ffprobePath, err := exec.LookPath("ffprobe")
	if err != nil {
		return false, false
	}
	cmd := exec.Command(
		ffprobePath,
		"-v", "error",
		"-select_streams", "a",
		"-show_entries", "stream=index",
		"-of", "csv=p=0",
		inputPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, false
	}
	return strings.TrimSpace(string(out)) != "", true
}

func isValidExtractAudioFormat(format string) bool {
	switch format {
	case "mp3", "wav", "ogg", "flac", "aac", "m4a", "opus":
		return true
	default:
		return false
	}
}

func extractAudioCodecArgs(targetFormat string, quality int, copyMode bool) []string {
	if copyMode {
		return []string{"-c:a", "copy"}
	}

	bitrate := "192k"
	if quality > 0 {
		switch {
		case quality <= 25:
			bitrate = "96k"
		case quality <= 50:
			bitrate = "128k"
		case quality <= 75:
			bitrate = "192k"
		default:
			bitrate = "320k"
		}
	}

	switch targetFormat {
	case "mp3":
		return []string{"-c:a", "libmp3lame", "-b:a", bitrate}
	case "wav":
		return []string{"-c:a", "pcm_s16le"}
	case "ogg":
		return []string{"-c:a", "libvorbis", "-b:a", bitrate}
	case "flac":
		return []string{"-c:a", "flac"}
	case "aac", "m4a":
		return []string{"-c:a", "aac", "-b:a", bitrate}
	case "opus":
		return []string{"-c:a", "libopus", "-b:a", bitrate}
	default:
		return []string{"-b:a", bitrate}
	}
}
