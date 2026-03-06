package ai

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestTranscribeMediaSuccess(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.wav")
	if err := os.WriteFile(input, []byte("fake-audio"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/audio/transcriptions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("missing auth header, got: %s", got)
		}
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			t.Fatalf("multipart parse failed: %v", err)
		}
		if r.FormValue("model") == "" {
			t.Fatalf("model field is required")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"text":"Merhaba dunya"}`))
	}))
	defer server.Close()

	gw := NewGateway(Policy{AllowedRoots: []string{root}})
	result, err := gw.TranscribeMedia(TranscribeMediaRequest{
		InputPath: input,
		Provider:  aiProviderOpenAICompatible,
		BaseURL:   server.URL,
		APIKey:    "test-key",
	})
	if err != nil {
		t.Fatalf("TranscribeMedia failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected %s, got %s", StatusSuccess, result.Status)
	}
	if strings.TrimSpace(result.TranscriptPath) == "" {
		t.Fatalf("expected transcript path")
	}
	data, readErr := os.ReadFile(result.TranscriptPath)
	if readErr != nil {
		t.Fatalf("read transcript failed: %v", readErr)
	}
	if !strings.Contains(string(data), "Merhaba dunya") {
		t.Fatalf("unexpected transcript content: %s", string(data))
	}
}

func TestSummarizeTranscriptSuccess(t *testing.T) {
	root := t.TempDir()
	transcript := filepath.Join(root, "meeting_transcript.txt")
	if err := os.WriteFile(transcript, []byte("Bugun urun yol haritasi konusuldu."), 0644); err != nil {
		t.Fatalf("write transcript failed: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("missing auth header, got: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"Kisa ozet metni"}}]}`))
	}))
	defer server.Close()

	gw := NewGateway(Policy{AllowedRoots: []string{root}})
	result, err := gw.SummarizeTranscript(SummarizeTranscriptRequest{
		TranscriptPath: transcript,
		Provider:       aiProviderOpenAICompatible,
		BaseURL:        server.URL,
		APIKey:         "test-key",
	})
	if err != nil {
		t.Fatalf("SummarizeTranscript failed: %v", err)
	}
	if result.Status != StatusSuccess {
		t.Fatalf("expected %s, got %s", StatusSuccess, result.Status)
	}
	if strings.TrimSpace(result.SummaryPath) == "" {
		t.Fatalf("expected summary path")
	}
	data, readErr := os.ReadFile(result.SummaryPath)
	if readErr != nil {
		t.Fatalf("read summary failed: %v", readErr)
	}
	if !strings.Contains(string(data), "Kisa ozet metni") {
		t.Fatalf("unexpected summary content: %s", string(data))
	}
}
