package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandUsesDocufyName(t *testing.T) {
	if rootCmd.Use != "docufy" {
		t.Fatalf("expected root command use to be docufy, got %q", rootCmd.Use)
	}

	var help bytes.Buffer
	rootCmd.SetOut(&help)
	rootCmd.SetErr(&help)
	if err := rootCmd.Help(); err != nil {
		t.Fatalf("Help failed: %v", err)
	}

	output := help.String()
	if strings.Contains(output, "fileconverter-cli") {
		t.Fatalf("help output still contains old command name: %s", output)
	}
	if !strings.Contains(output, "docufy") {
		t.Fatalf("help output does not contain docufy: %s", output)
	}
}

func TestZshCompletionUsesDocufyName(t *testing.T) {
	var completion bytes.Buffer
	if err := rootCmd.GenZshCompletion(&completion); err != nil {
		t.Fatalf("GenZshCompletion failed: %v", err)
	}

	output := completion.String()
	if !strings.Contains(output, "_docufy") {
		t.Fatalf("expected zsh completion to contain _docufy")
	}
	if strings.Contains(output, "_fileconverter-cli") {
		t.Fatalf("completion output still contains old completion name")
	}
}
