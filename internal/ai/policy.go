package ai

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Policy AI tool çağrılarının dosya erişim sınırlarını belirler.
type Policy struct {
	AllowedRoots []string
}

func (p Policy) ResolvePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path zorunlu")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func (p Policy) ValidatePath(path string) (string, error) {
	abs, err := p.ResolvePath(path)
	if err != nil {
		return "", err
	}
	if len(p.AllowedRoots) == 0 {
		return abs, nil
	}

	for _, root := range p.AllowedRoots {
		rootAbs, rootErr := filepath.Abs(root)
		if rootErr != nil {
			continue
		}
		if sameOrSubPath(abs, rootAbs) {
			return abs, nil
		}
	}
	return "", fmt.Errorf("path policy disi: %s", abs)
}

func sameOrSubPath(path string, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	rootWithSep := root + string(filepath.Separator)
	return strings.HasPrefix(path, rootWithSep)
}
