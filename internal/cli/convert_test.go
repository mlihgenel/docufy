package cli

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/mlihgenel/docufy/v2/internal/config"
	"github.com/mlihgenel/docufy/v2/internal/converter"
)

func TestConvertCreatesMissingOutputDir(t *testing.T) {
	prevOutputDir := outputDir
	prevOutputFormat := outputFormat
	prevVerbose := verbose
	prevActiveProjectConfig := activeProjectConfig
	prevActiveProjectConfigPath := activeProjectConfigPath
	prevToFormat := toFormat
	prevConvertProfile := convertProfile
	prevQuality := quality
	prevCustomName := customName
	prevConvertOnConflict := convertOnConflict
	prevConvertPreserveMD := convertPreserveMD
	prevConvertStripMD := convertStripMD
	prevConvertPreset := convertPreset
	prevConvertWidth := convertWidth
	prevConvertHeight := convertHeight
	prevConvertUnit := convertUnit
	prevConvertResizeDPI := convertResizeDPI
	prevConvertResizeMode := convertResizeMode
	prevConvertOptimize := convertOptimize
	prevConvertTargetSize := convertTargetSize
	t.Cleanup(func() {
		outputDir = prevOutputDir
		outputFormat = prevOutputFormat
		verbose = prevVerbose
		activeProjectConfig = prevActiveProjectConfig
		activeProjectConfigPath = prevActiveProjectConfigPath
		toFormat = prevToFormat
		convertProfile = prevConvertProfile
		quality = prevQuality
		customName = prevCustomName
		convertOnConflict = prevConvertOnConflict
		convertPreserveMD = prevConvertPreserveMD
		convertStripMD = prevConvertStripMD
		convertPreset = prevConvertPreset
		convertWidth = prevConvertWidth
		convertHeight = prevConvertHeight
		convertUnit = prevConvertUnit
		convertResizeDPI = prevConvertResizeDPI
		convertResizeMode = prevConvertResizeMode
		convertOptimize = prevConvertOptimize
		convertTargetSize = prevConvertTargetSize
	})

	for _, env := range []string{
		envOutput,
		legacyEnvOutput,
		envQuality,
		legacyEnvQuality,
		envProfile,
		legacyEnvProfile,
		envConflict,
		legacyEnvConflict,
		envMetadata,
		legacyEnvMetadata,
	} {
		t.Setenv(env, "")
	}

	activeProjectConfig = &config.ProjectConfig{}
	activeProjectConfigPath = ""
	verbose = false
	outputFormat = OutputFormatJSON
	outputDir = filepath.Join(t.TempDir(), "nested", "out")
	toFormat = "webp"
	convertProfile = ""
	quality = 0
	customName = ""
	convertOnConflict = converter.ConflictVersioned
	convertPreserveMD = false
	convertStripMD = false
	convertPreset = ""
	convertWidth = 0
	convertHeight = 0
	convertUnit = ""
	convertResizeDPI = 0
	convertResizeMode = ""
	convertOptimize = false
	convertTargetSize = ""

	input := filepath.Join(t.TempDir(), "input.png")
	if err := writeTestPNG(input); err != nil {
		t.Fatalf("write test png failed: %v", err)
	}

	if err := convertCmd.RunE(convertCmd, []string{input}); err != nil {
		t.Fatalf("convert command failed: %v", err)
	}

	output := filepath.Join(outputDir, "input.webp")
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
}

func writeTestPNG(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 32, G: 180, B: 160, A: 255})
		}
	}

	return png.Encode(f, img)
}
