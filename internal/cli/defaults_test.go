package cli

import (
	"os"
	"testing"

	"github.com/mlihgenel/docufy/internal/config"
	"github.com/spf13/cobra"
)

func TestApplyRootDefaultsEnvOverridesConfig(t *testing.T) {
	prevCfg := activeProjectConfig
	defer func() { activeProjectConfig = prevCfg }()

	activeProjectConfig = &config.ProjectConfig{
		DefaultOutput: "/from-config",
		Workers:       3,
	}
	outputDir = ""
	workers = 1

	t.Setenv(envOutput, "/from-env")
	t.Setenv(envWorkers, "9")

	c := newTestRootCommand()
	if err := applyRootDefaults(c); err != nil {
		t.Fatalf("applyRootDefaults failed: %v", err)
	}

	if outputDir != "/from-env" {
		t.Fatalf("expected env output, got %s", outputDir)
	}
	if workers != 9 {
		t.Fatalf("expected env workers 9, got %d", workers)
	}
}

func TestApplyRootDefaultsRespectsChangedFlags(t *testing.T) {
	prevCfg := activeProjectConfig
	defer func() { activeProjectConfig = prevCfg }()

	activeProjectConfig = &config.ProjectConfig{
		DefaultOutput: "/from-config",
		Workers:       5,
	}
	outputDir = "/manual"
	workers = 11

	c := newTestRootCommand()
	if err := c.Flags().Set("output", "/manual"); err != nil {
		t.Fatalf("set output flag failed: %v", err)
	}
	if err := c.Flags().Set("workers", "11"); err != nil {
		t.Fatalf("set workers flag failed: %v", err)
	}

	if err := applyRootDefaults(c); err != nil {
		t.Fatalf("applyRootDefaults failed: %v", err)
	}

	if outputDir != "/manual" {
		t.Fatalf("expected manual output unchanged, got %s", outputDir)
	}
	if workers != 11 {
		t.Fatalf("expected manual workers unchanged, got %d", workers)
	}
}

func newTestRootCommand() *cobra.Command {
	c := &cobra.Command{Use: "test"}
	c.Flags().String("output", "", "")
	c.Flags().Int("workers", 0, "")
	return c
}

func TestReadEnvHelpers(t *testing.T) {
	t.Setenv("X_INT", "12")
	if v, ok := readEnvInt("X_INT"); !ok || v != 12 {
		t.Fatalf("unexpected int parse result")
	}

	t.Setenv("X_DUR", "2s")
	if _, ok := readEnvDuration("X_DUR"); !ok {
		t.Fatalf("expected duration parse success")
	}

	_ = os.Unsetenv("X_INT")
	_ = os.Unsetenv("X_DUR")
}

func TestApplyRootDefaultsLegacyEnvFallback(t *testing.T) {
	prevCfg := activeProjectConfig
	defer func() { activeProjectConfig = prevCfg }()

	activeProjectConfig = &config.ProjectConfig{
		DefaultOutput: "/from-config",
		Workers:       3,
	}
	outputDir = ""
	workers = 1

	t.Setenv(legacyEnvOutput, "/from-legacy")
	t.Setenv(legacyEnvWorkers, "7")

	c := newTestRootCommand()
	if err := applyRootDefaults(c); err != nil {
		t.Fatalf("applyRootDefaults failed: %v", err)
	}

	if outputDir != "/from-legacy" {
		t.Fatalf("expected legacy env output, got %s", outputDir)
	}
	if workers != 7 {
		t.Fatalf("expected legacy env workers 7, got %d", workers)
	}
}

func TestApplyRootDefaultsPrefersDocufyEnvOverLegacy(t *testing.T) {
	prevCfg := activeProjectConfig
	defer func() { activeProjectConfig = prevCfg }()

	activeProjectConfig = nil
	outputDir = ""
	workers = 1

	t.Setenv(envOutput, "/from-docufy")
	t.Setenv(legacyEnvOutput, "/from-legacy")
	t.Setenv(envWorkers, "9")
	t.Setenv(legacyEnvWorkers, "7")

	c := newTestRootCommand()
	if err := applyRootDefaults(c); err != nil {
		t.Fatalf("applyRootDefaults failed: %v", err)
	}

	if outputDir != "/from-docufy" {
		t.Fatalf("expected docufy env output, got %s", outputDir)
	}
	if workers != 9 {
		t.Fatalf("expected docufy env workers 9, got %d", workers)
	}
}
