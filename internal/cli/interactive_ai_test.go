package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aigateway "github.com/mlihgenel/docufy/v2/internal/ai"
	"github.com/mlihgenel/docufy/v2/internal/converter"
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

func TestAICommandDoneTransitionsToDoneState(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.state = stateAIExecuting
	m.aiPendingPrompt = "dosya bilgisi ver"

	nextModel, cmd := m.Update(aiToolDoneMsg{resultText: "işlem tamamlandı", currentFile: "/tmp/a.mp4"})
	if cmd != nil {
		t.Fatalf("expected no command after aiToolDoneMsg")
	}
	next, ok := nextModel.(interactiveModel)
	if !ok {
		t.Fatalf("unexpected model type")
	}
	if next.state != stateAIDone {
		t.Fatalf("expected stateAIDone, got %v", next.state)
	}
	if next.aiLastResult != "işlem tamamlandı" {
		t.Fatalf("unexpected aiLastResult: %q", next.aiLastResult)
	}
	if next.aiPendingPrompt != "" {
		t.Fatalf("expected pending prompt to be cleared")
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

func TestParseAIOnConflictPolicy(t *testing.T) {
	if got := parseAIOnConflictPolicy("dosyayı overwrite et"); got != converter.ConflictOverwrite {
		t.Fatalf("expected overwrite, got %q", got)
	}
	if got := parseAIOnConflictPolicy("dosyayı atla ve skip kullan"); got != converter.ConflictSkip {
		t.Fatalf("expected skip, got %q", got)
	}
	if got := parseAIOnConflictPolicy("normal dönüştür"); got != converter.ConflictVersioned {
		t.Fatalf("expected versioned, got %q", got)
	}
}

func TestAINeedsHighRiskAckUsesParsedPolicy(t *testing.T) {
	if !aiNeedsHighRiskAck("bu dosyayı png yap ve overwrite et") {
		t.Fatalf("expected high risk ack for overwrite policy")
	}
	if aiNeedsHighRiskAck("bu dosyayı png yap ve skip kullan") {
		t.Fatalf("did not expect high risk ack for skip policy")
	}
}

func TestExecuteAICommandTrimRemoveUsesRemoveMode(t *testing.T) {
	root := t.TempDir()
	input := filepath.Join(root, "sample.mp4")
	if err := os.WriteFile(input, []byte("fake"), 0644); err != nil {
		t.Fatalf("write input failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	_, _, err := executeAICommand(gw, "videoyu kırp ve 5 ile 7 saniyeyi sil", input)
	if !errors.Is(err, aigateway.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented for remove mode, got %v", err)
	}
}

func TestParseAIFormatPairFromTo(t *testing.T) {
	from, to := parseAIFormatPair("from mp4 to gif")
	if from != "mp4" || to != "gif" {
		t.Fatalf("unexpected format pair: from=%q to=%q", from, to)
	}

	from, to = parseAIFormatPair("png'den webp'ye toplu dönüştür")
	if from != "png" || to != "webp" {
		t.Fatalf("unexpected tr format pair: from=%q to=%q", from, to)
	}
}

func TestExecuteAICommandWatchCreatesTaskFile(t *testing.T) {
	root := t.TempDir()
	watchDir := filepath.Join(root, "incoming")
	if err := os.MkdirAll(watchDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	result, _, err := executeAICommand(gw, "watch "+watchDir+" from mp4 to gif", "")
	if err != nil {
		t.Fatalf("executeAICommand failed: %v", err)
	}
	if !strings.Contains(result, "Watch görevi üretildi") {
		t.Fatalf("unexpected result: %q", result)
	}

	taskDir := filepath.Join(root, ".docufy", "tasks")
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		t.Fatalf("read task dir failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one generated task file")
	}
}

func TestExecuteAICommandBatchCreatesTaskFile(t *testing.T) {
	root := t.TempDir()
	batchDir := filepath.Join(root, "images")
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	gw := aigateway.NewGateway(aigateway.Policy{AllowedRoots: []string{root}})
	result, _, err := executeAICommand(gw, "toplu "+batchDir+" from png to webp", "")
	if err != nil {
		t.Fatalf("executeAICommand failed: %v", err)
	}
	if !strings.Contains(result, "Batch görevi üretildi") {
		t.Fatalf("unexpected result: %q", result)
	}

	taskDir := filepath.Join(root, ".docufy", "tasks")
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		t.Fatalf("read task dir failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("expected at least one generated task file")
	}
}
