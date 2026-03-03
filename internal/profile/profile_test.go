package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveBuiltins(t *testing.T) {
	tests := []string{"social-story", "podcast-clean", "archive-lossless"}
	for _, name := range tests {
		p, err := Resolve(name)
		if err != nil {
			t.Fatalf("Resolve(%s) failed: %v", name, err)
		}
		if p.Name == "" {
			t.Fatalf("Resolve(%s) returned empty profile name", name)
		}
	}
}

func TestResolveInvalid(t *testing.T) {
	_, err := Resolve("unknown-profile")
	if err == nil {
		t.Fatalf("expected error for unknown profile")
	}
}

func TestListMergesBuiltinAndUserProfiles(t *testing.T) {
	tempHome := t.TempDir()
	restore := withUserHomeDir(tempHome)
	defer restore()

	profilesDir := filepath.Join(tempHome, ".docufy", "profiles")
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	override := strings.Join([]string{
		`name = "social-story"`,
		`description = "Kullanici override"`,
		`quality = 71`,
		`metadata_mode = "preserve"`,
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(profilesDir, "social-story.toml"), []byte(override), 0644); err != nil {
		t.Fatalf("WriteFile override failed: %v", err)
	}

	custom := strings.Join([]string{
		`name = "print-a4"`,
		`description = "A4 cikti profili"`,
		`width = 210`,
		`height = 297`,
		`unit = "mm"`,
		`dpi = 300`,
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(profilesDir, "print-a4.toml"), []byte(custom), 0644); err != nil {
		t.Fatalf("WriteFile custom failed: %v", err)
	}

	items, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) < 4 {
		t.Fatalf("expected builtin + user profiles, got %d", len(items))
	}

	merged, err := Resolve("social-story")
	if err != nil {
		t.Fatalf("Resolve merged profile failed: %v", err)
	}
	if merged.Source != "builtin+user" {
		t.Fatalf("expected merged source, got %q", merged.Source)
	}
	if merged.Description != "Kullanici override" {
		t.Fatalf("expected override description, got %q", merged.Description)
	}
	if merged.Quality == nil || *merged.Quality != 71 {
		t.Fatalf("expected override quality 71, got %#v", merged.Quality)
	}

	userProfile, err := Resolve("print-a4")
	if err != nil {
		t.Fatalf("Resolve user profile failed: %v", err)
	}
	if userProfile.Source != "user" {
		t.Fatalf("expected user source, got %q", userProfile.Source)
	}
	if userProfile.Width == nil || *userProfile.Width != 210 {
		t.Fatalf("expected width 210, got %#v", userProfile.Width)
	}
}

func TestSaveUserProfileWritesToml(t *testing.T) {
	tempHome := t.TempDir()
	restore := withUserHomeDir(tempHome)
	defer restore()

	path, err := SaveUserProfile(Definition{
		Name:         "My-Profile",
		Description:  "Test profili",
		Quality:      IntPtr(88),
		Retry:        IntPtr(2),
		RetryDelay:   DurationPtr(2 * time.Second),
		ResizePreset: "story",
		MetadataMode: "strip",
	})
	if err != nil {
		t.Fatalf("SaveUserProfile failed: %v", err)
	}

	if want := filepath.Join(tempHome, ".docufy", "profiles", "my-profile.toml"); path != want {
		t.Fatalf("unexpected path: got %q want %q", path, want)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	text := string(content)
	for _, snippet := range []string{
		`name = "my-profile"`,
		`description = "Test profili"`,
		`quality = 88`,
		`retry = 2`,
		`retry_delay = "2s"`,
		`resize_preset = "story"`,
		`metadata_mode = "strip"`,
	} {
		if !strings.Contains(text, snippet) {
			t.Fatalf("expected %q in file content: %s", snippet, text)
		}
	}
}

func TestNamesCount(t *testing.T) {
	names := Names()
	if len(names) < 3 {
		t.Fatalf("expected at least 3 built-in profiles")
	}
}

func withUserHomeDir(home string) func() {
	prev := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	return func() {
		userHomeDir = prev
	}
}

func TestListFallsBackToLegacyProfiles(t *testing.T) {
	tempHome := t.TempDir()
	restore := withUserHomeDir(tempHome)
	defer restore()

	legacyDir := filepath.Join(tempHome, ".fileconverter", "profiles")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(legacyDir, "legacy-only.toml"), []byte("name = \"legacy-only\"\nquality = 55\n"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	p, err := Resolve("legacy-only")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if p.Source != "user" {
		t.Fatalf("expected legacy profile to load as user profile, got %q", p.Source)
	}
	if p.Quality == nil || *p.Quality != 55 {
		t.Fatalf("expected quality 55, got %#v", p.Quality)
	}
}

func TestDocufyProfilesOverrideLegacyProfiles(t *testing.T) {
	tempHome := t.TempDir()
	restore := withUserHomeDir(tempHome)
	defer restore()

	legacyDir := filepath.Join(tempHome, ".fileconverter", "profiles")
	newDir := filepath.Join(tempHome, ".docufy", "profiles")
	if err := os.MkdirAll(legacyDir, 0755); err != nil {
		t.Fatalf("MkdirAll legacy failed: %v", err)
	}
	if err := os.MkdirAll(newDir, 0755); err != nil {
		t.Fatalf("MkdirAll new failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(legacyDir, "social-story.toml"), []byte("name = \"social-story\"\nquality = 61\n"), 0644); err != nil {
		t.Fatalf("WriteFile legacy failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "social-story.toml"), []byte("name = \"social-story\"\nquality = 77\n"), 0644); err != nil {
		t.Fatalf("WriteFile new failed: %v", err)
	}

	p, err := Resolve("social-story")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if p.Quality == nil || *p.Quality != 77 {
		t.Fatalf("expected docufy profile to win with quality 77, got %#v", p.Quality)
	}
}
