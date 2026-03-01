package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/mlihgenel/fileconverter-cli/internal/profile"
)

func TestBuildProfileDefinitionFromInputNonInteractive(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(bytes.NewBufferString(""))
	cmd.SetOut(&bytes.Buffer{})

	cmd.Flags().Int("quality", 0, "")
	cmd.Flags().String("on-conflict", "", "")
	cmd.Flags().Int("retry", 0, "")
	cmd.Flags().String("retry-delay", "", "")
	cmd.Flags().String("report", "", "")
	cmd.Flags().String("preset", "", "")
	cmd.Flags().String("resize-mode", "", "")
	cmd.Flags().Float64("width", 0, "")
	cmd.Flags().Float64("height", 0, "")
	cmd.Flags().String("unit", "", "")
	cmd.Flags().Float64("dpi", 0, "")
	cmd.Flags().String("metadata-mode", "", "")

	t.Cleanup(resetProfileCreateFlags)
	profileCreateDescription = "Test"
	profileCreateQuality = 83
	profileCreateOnConflict = "skip"
	profileCreateRetry = 2
	profileCreateRetryDelay = "750ms"
	profileCreateReport = "json"
	profileCreatePreset = "story"
	profileCreateResizeMode = "fit"
	profileCreateWidth = 1080
	profileCreateHeight = 1920
	profileCreateUnit = "px"
	profileCreateDPI = 300
	profileCreateMetadataMode = "strip"

	if err := cmd.Flags().Set("quality", "83"); err != nil {
		t.Fatalf("Set quality failed: %v", err)
	}
	if err := cmd.Flags().Set("retry", "2"); err != nil {
		t.Fatalf("Set retry failed: %v", err)
	}
	if err := cmd.Flags().Set("width", "1080"); err != nil {
		t.Fatalf("Set width failed: %v", err)
	}
	if err := cmd.Flags().Set("height", "1920"); err != nil {
		t.Fatalf("Set height failed: %v", err)
	}
	if err := cmd.Flags().Set("dpi", "300"); err != nil {
		t.Fatalf("Set dpi failed: %v", err)
	}

	def, err := buildProfileDefinitionFromInput(cmd, []string{"story-fast"})
	if err != nil {
		t.Fatalf("buildProfileDefinitionFromInput failed: %v", err)
	}

	if def.Name != "story-fast" {
		t.Fatalf("expected name story-fast, got %q", def.Name)
	}
	if def.Quality == nil || *def.Quality != 83 {
		t.Fatalf("expected quality 83, got %#v", def.Quality)
	}
	if def.RetryDelay == nil || *def.RetryDelay != 750*time.Millisecond {
		t.Fatalf("expected retry delay 750ms, got %#v", def.RetryDelay)
	}
	if def.Width == nil || *def.Width != 1080 {
		t.Fatalf("expected width 1080, got %#v", def.Width)
	}
	if def.MetadataMode != "strip" {
		t.Fatalf("expected metadata mode strip, got %q", def.MetadataMode)
	}
}

func TestSettingsProfileLinesShowsUserProfiles(t *testing.T) {
	tempHome := t.TempDir()
	prevHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tempHome); err != nil {
		t.Fatalf("Setenv failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("HOME", prevHome)
	}()

	_, err := profile.SaveUserProfile(profile.Definition{
		Name:        "ops-safe",
		Description: "Ops varsayilanlari",
		Quality:     profile.IntPtr(91),
	})
	if err != nil {
		t.Fatalf("SaveUserProfile failed: %v", err)
	}

	lines := settingsProfileLines()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "social-story (builtin)") {
		t.Fatalf("expected builtin profile in settings: %s", joined)
	}
	if !strings.Contains(joined, "ops-safe (user)") {
		t.Fatalf("expected user profile in settings: %s", joined)
	}
}

func resetProfileCreateFlags() {
	profileCreateDescription = ""
	profileCreateQuality = 0
	profileCreateOnConflict = ""
	profileCreateRetry = 0
	profileCreateRetryDelay = ""
	profileCreateReport = ""
	profileCreatePreset = ""
	profileCreateResizeMode = ""
	profileCreateWidth = 0
	profileCreateHeight = 0
	profileCreateUnit = ""
	profileCreateDPI = 0
	profileCreateMetadataMode = ""
}
