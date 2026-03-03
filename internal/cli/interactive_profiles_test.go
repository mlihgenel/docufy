package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/mlihgenel/docufy/v2/internal/profile"
)

func TestGoToProfileSelectIncludesProfileChoices(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.selectedCategory = 3
	m.sourceFormat = "mp4"
	m.targetFormat = "gif"

	next := m.goToProfileSelect(false)
	if next.state != stateSelectProfile {
		t.Fatalf("expected stateSelectProfile, got %v", next.state)
	}
	if len(next.choices) < 2 {
		t.Fatalf("expected profile choices, got %v", next.choices)
	}
	if next.choices[0] != "Profil kullanma" {
		t.Fatalf("expected no-profile option first, got %q", next.choices[0])
	}

	joined := strings.Join(next.choices, "\n")
	if !strings.Contains(joined, "social-story") {
		t.Fatalf("expected built-in profile in choices: %s", joined)
	}
}

func TestGoBackFromFileBrowserReturnsProfileSelect(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.selectedCategory = 3
	m.sourceFormat = "mp4"
	m.targetFormat = "gif"
	m.state = stateFileBrowser

	next := m.goBack()
	if next.state != stateSelectProfile {
		t.Fatalf("expected stateSelectProfile, got %v", next.state)
	}
}

func TestInteractiveProfileOverridesEffectiveSettings(t *testing.T) {
	m := newInteractiveModel(nil, false)
	m.defaultQuality = 10
	m.defaultOnConflict = "versioned"
	m.defaultRetry = 0
	m.defaultRetryDelay = 500 * time.Millisecond
	m.defaultReport = "off"
	m.profileActive = true
	m.selectedProfile = "story-fast"
	m.profileValue = profile.Definition{
		Name:         "story-fast",
		Quality:      profile.IntPtr(91),
		OnConflict:   "skip",
		Retry:        profile.IntPtr(3),
		RetryDelay:   profile.DurationPtr(2 * time.Second),
		Report:       "json",
		MetadataMode: "strip",
	}

	if got := m.effectiveQuality(); got != 91 {
		t.Fatalf("expected quality 91, got %d", got)
	}
	if got := m.effectiveOnConflict(); got != "skip" {
		t.Fatalf("expected conflict skip, got %q", got)
	}
	if got := m.effectiveRetry(); got != 3 {
		t.Fatalf("expected retry 3, got %d", got)
	}
	if got := m.effectiveRetryDelay(); got != 2*time.Second {
		t.Fatalf("expected retry delay 2s, got %s", got)
	}
	if got := m.effectiveReport(); got != "json" {
		t.Fatalf("expected report json, got %q", got)
	}
	if got := m.effectiveMetadataMode(); got != "strip" {
		t.Fatalf("expected metadata strip, got %q", got)
	}
}
