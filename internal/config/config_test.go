package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFallsBackToLegacyPath(t *testing.T) {
	tempHome := t.TempDir()
	restore := withConfigUserHomeDir(tempHome)
	defer restore()

	legacyDir := filepath.Join(tempHome, legacyConfigDirName)
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	content := `{"first_run_completed":true,"default_output_dir":"legacy-out"}`
	if err := os.WriteFile(filepath.Join(legacyDir, "config.json"), []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if !cfg.FirstRunCompleted {
		t.Fatalf("expected legacy first_run_completed to load")
	}
	if cfg.DefaultOutputDir != "legacy-out" {
		t.Fatalf("unexpected output dir: %q", cfg.DefaultOutputDir)
	}
}

func TestLoadConfigPrefersDocufyPath(t *testing.T) {
	tempHome := t.TempDir()
	restore := withConfigUserHomeDir(tempHome)
	defer restore()

	newDir := filepath.Join(tempHome, configDirName)
	legacyDir := filepath.Join(tempHome, legacyConfigDirName)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("MkdirAll new failed: %v", err)
	}
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("MkdirAll legacy failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "config.json"), []byte(`{"default_output_dir":"legacy-out"}`), 0644); err != nil {
		t.Fatalf("WriteFile legacy failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "config.json"), []byte(`{"default_output_dir":"docufy-out"}`), 0644); err != nil {
		t.Fatalf("WriteFile new failed: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.DefaultOutputDir != "docufy-out" {
		t.Fatalf("expected docufy config to win, got %q", cfg.DefaultOutputDir)
	}
}

func TestSaveConfigWritesDocufyPath(t *testing.T) {
	tempHome := t.TempDir()
	restore := withConfigUserHomeDir(tempHome)
	defer restore()

	if err := SaveConfig(&AppConfig{FirstRunCompleted: true, DefaultOutputDir: "docufy-out"}); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	newPath := filepath.Join(tempHome, configDirName, "config.json")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("expected new config path to exist: %v", err)
	}
	legacyPath := filepath.Join(tempHome, legacyConfigDirName, "config.json")
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("expected legacy path to remain untouched, got err=%v", err)
	}
}

func TestSetAndGetAISettings(t *testing.T) {
	tempHome := t.TempDir()
	restore := withConfigUserHomeDir(tempHome)
	defer restore()

	expected := AISettings{
		Provider:   "openai",
		Model:      "gpt-4o-mini",
		BaseURL:    "https://api.openai.com/v1",
		SidecarURL: "http://127.0.0.1:8081",
	}
	if err := SetAISettings(expected); err != nil {
		t.Fatalf("SetAISettings failed: %v", err)
	}

	got := GetAISettings()
	if got.Provider != expected.Provider {
		t.Fatalf("provider mismatch: got=%q want=%q", got.Provider, expected.Provider)
	}
	if got.Model != expected.Model {
		t.Fatalf("model mismatch: got=%q want=%q", got.Model, expected.Model)
	}
	if got.BaseURL != expected.BaseURL {
		t.Fatalf("base_url mismatch: got=%q want=%q", got.BaseURL, expected.BaseURL)
	}
	if got.SidecarURL != expected.SidecarURL {
		t.Fatalf("sidecar_url mismatch: got=%q want=%q", got.SidecarURL, expected.SidecarURL)
	}
}

func withConfigUserHomeDir(home string) func() {
	prev := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	return func() {
		userHomeDir = prev
	}
}
