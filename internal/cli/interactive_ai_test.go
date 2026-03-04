package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	aigateway "github.com/mlihgenel/docufy/v2/internal/ai"
)

func TestExecuteAICommandSetCurrentFile(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(input, []byte("hello"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	result, current, err := executeAICommand(gw, "/dosya "+input, "")
	if err != nil {
		t.Fatalf("executeAICommand failed: %v", err)
	}
	if current != input {
		t.Fatalf("current file mismatch: got=%q want=%q", current, input)
	}
	if !strings.Contains(result, "Aktif dosya ayarlandı") {
		t.Fatalf("unexpected result text: %q", result)
	}
}

func TestExecuteAICommandInfoWithCurrentFile(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(input, []byte("hello"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	result, current, err := executeAICommand(gw, "dosya bilgisi ver", input)
	if err != nil {
		t.Fatalf("executeAICommand failed: %v", err)
	}
	if current != input {
		t.Fatalf("current file mismatch: got=%q want=%q", current, input)
	}
	if !strings.Contains(result, "Dosya:") {
		t.Fatalf("unexpected result text: %q", result)
	}
}

func TestExecuteAICommandUnknownIntent(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.txt")
	if err := os.WriteFile(input, []byte("hello"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	result, _, err := executeAICommand(gw, "bu komutu anlamsizca dene", input)
	if err != nil {
		t.Fatalf("executeAICommand failed: %v", err)
	}
	if !strings.Contains(result, "Komut anlaşılamadı") {
		t.Fatalf("unexpected result text: %q", result)
	}
}
