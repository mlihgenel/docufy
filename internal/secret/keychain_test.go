package secret

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeProvider(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "openai", want: "openai"},
		{in: "OPENAI", want: "openai"},
		{in: "openai-compatible", want: "openai-compatible"},
		{in: "openai_compatible", want: "openai-compatible"},
		{in: "ollama", want: "ollama"},
		{in: "unknown", want: ""},
	}

	for _, tc := range tests {
		got := normalizeProvider(tc.in)
		if got != tc.want {
			t.Fatalf("normalizeProvider(%q)=%q want=%q", tc.in, got, tc.want)
		}
	}
}

func TestAIKeychainAccount(t *testing.T) {
	got := aiKeychainAccount("openai")
	if got != "ai:openai" {
		t.Fatalf("unexpected account: %q", got)
	}
}

func TestKeychainDisabled(t *testing.T) {
	t.Setenv("DOCUFY_AI_DISABLE_KEYCHAIN", "1")

	if err := SaveAIAPIKey("openai", "secret"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled from SaveAIAPIKey, got %v", err)
	}

	key, err := LoadAIAPIKey("openai")
	if !errors.Is(err, ErrDisabled) {
		t.Fatalf("expected ErrDisabled from LoadAIAPIKey, got %v", err)
	}
	if key != "" {
		t.Fatalf("expected empty key when disabled, got %q", key)
	}
}

func TestPsSingleQuote(t *testing.T) {
	got := psSingleQuote("a'b'c")
	if got != "a''b''c" {
		t.Fatalf("unexpected psSingleQuote output: %q", got)
	}
}

func TestWindowsSecretFilePathSanitizesAccount(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	path, err := windowsSecretFilePath("ai:openai")
	if err != nil {
		t.Fatalf("windowsSecretFilePath failed: %v", err)
	}
	if path == "" {
		t.Fatalf("expected non-empty path")
	}
	if !strings.HasSuffix(path, "ai_openai.dpapi") {
		t.Fatalf("unexpected file suffix: %q", path)
	}
}
