package ai

import (
	"path/filepath"
	"testing"
)

func TestPolicyValidatePathWithinRoot(t *testing.T) {
	root := t.TempDir()
	p := Policy{AllowedRoots: []string{root}}

	target := filepath.Join(root, "nested", "a.txt")
	got, err := p.ValidatePath(target)
	if err != nil {
		t.Fatalf("ValidatePath failed: %v", err)
	}
	if got != target {
		t.Fatalf("expected %s, got %s", target, got)
	}
}

func TestPolicyValidatePathOutsideRoot(t *testing.T) {
	root := t.TempDir()
	p := Policy{AllowedRoots: []string{root}}

	otherRoot := t.TempDir()
	target := filepath.Join(otherRoot, "a.txt")
	if _, err := p.ValidatePath(target); err == nil {
		t.Fatalf("expected policy error for outside path")
	}
}
