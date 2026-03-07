package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	// converter kayıtları için side-effect import
	_ "github.com/mlihgenel/docufy/v2/internal/converter"
)

func TestExecuteConvertPipeline(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(input, []byte("hello pipeline"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	spec := Spec{
		Input: input,
		Steps: []Step{
			{Type: StepConvert, To: "md"},
		},
	}

	result, err := Execute(spec, ExecuteConfig{
		OutputDir:      dir,
		DefaultQuality: 80,
		MetadataMode:   "auto",
		OnConflict:     "versioned",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result.Steps) != 1 {
		t.Fatalf("unexpected step count: %d", len(result.Steps))
	}
	if !result.Steps[0].Success {
		t.Fatalf("step should be success: %#v", result.Steps[0])
	}
	if result.FinalOutput == "" {
		t.Fatalf("final output should not be empty")
	}
	if _, err := os.Stat(result.FinalOutput); err != nil {
		t.Fatalf("final output not found: %v", err)
	}
}

func TestExecuteInvalidSpec(t *testing.T) {
	_, err := Execute(Spec{}, ExecuteConfig{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestExecuteCreatesMissingOutputDir(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(input, []byte("hello pipeline"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	outputDir := filepath.Join(dir, "nested", "out")
	spec := Spec{
		Input: input,
		Steps: []Step{
			{Type: StepConvert, To: "md"},
		},
	}

	result, err := Execute(spec, ExecuteConfig{
		OutputDir:      outputDir,
		DefaultQuality: 80,
		MetadataMode:   "auto",
		OnConflict:     "versioned",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	want := filepath.Join(outputDir, "input.md")
	if result.FinalOutput != want {
		t.Fatalf("expected final output %s, got %s", want, result.FinalOutput)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected final output in nested output dir: %v", err)
	}
}

func TestExecuteExtractAudioPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found; skipping integration test")
	}

	dir := t.TempDir()
	input := filepath.Join(dir, "input.mp4")
	if err := generatePipelineTestVideo(input); err != nil {
		t.Fatalf("generate test video failed: %v", err)
	}

	outputDir := filepath.Join(dir, "nested", "audio")
	spec := Spec{
		Input:  input,
		Output: filepath.Join(outputDir, "final.mp3"),
		Steps: []Step{
			{Type: StepExtractAudio, To: "wav"},
			{Type: StepAudioNormalize},
			{Type: StepConvert, To: "mp3"},
		},
	}

	result, err := Execute(spec, ExecuteConfig{
		OutputDir:      outputDir,
		DefaultQuality: 80,
		MetadataMode:   "auto",
		OnConflict:     "versioned",
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(result.Steps) != 3 {
		t.Fatalf("unexpected step count: %d", len(result.Steps))
	}
	if result.FinalOutput != spec.Output {
		t.Fatalf("expected final output %s, got %s", spec.Output, result.FinalOutput)
	}
	if _, err := os.Stat(result.FinalOutput); err != nil {
		t.Fatalf("expected extracted+normalized output file: %v", err)
	}
}

func generatePipelineTestVideo(output string) error {
	args := []string{
		"-loglevel", "error",
		"-f", "lavfi",
		"-i", "color=c=blue:s=320x240:d=2",
		"-f", "lavfi",
		"-i", "sine=frequency=1000:duration=2",
		"-shortest",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-y",
		output,
	}
	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg test video generation failed: %s", string(out))
	}
	return nil
}
