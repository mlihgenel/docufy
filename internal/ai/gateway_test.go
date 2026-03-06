package ai

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestConvertFileRequiresTo(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(input, []byte("hello"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := NewGateway(Policy{AllowedRoots: []string{root}})
	_, err := gw.ConvertFile(ConvertFileRequest{
		InputPath: input,
	})
	if err == nil {
		t.Fatalf("expected error for missing target format")
	}
}

func TestTrimVideoRemoveNoLongerNotImplemented(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.mp4")
	if err := os.WriteFile(input, []byte("fake"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := NewGateway(Policy{AllowedRoots: []string{root}})
	result, err := gw.TrimVideo(TrimVideoRequest{
		InputPath: input,
		Mode:      "remove",
		Start:     "5",
		End:       "2",
	})
	if err == nil {
		t.Fatalf("expected validation error for invalid remove range")
	}
	if errors.Is(err, ErrNotImplemented) {
		t.Fatalf("did not expect ErrNotImplemented")
	}
	if result.Status == StatusNotImplemented {
		t.Fatalf("did not expect status %s", StatusNotImplemented)
	}
}

func TestGetFileInfoSuccess(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(input, []byte("hello"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := NewGateway(Policy{AllowedRoots: []string{root}})
	result, err := gw.GetFileInfo(FileInfoRequest{Path: input})
	if err != nil {
		t.Fatalf("GetFileInfo failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected status %s, got %s", StatusSuccess, result.Status)
	}
	if result.FileInfo == nil {
		t.Fatalf("expected non-nil file info")
	}
}
