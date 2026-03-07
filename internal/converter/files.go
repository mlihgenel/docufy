package converter

import (
	"os"
	"path/filepath"
)

// EnsureParentDir hedef dosyanın üst dizinini yoksa oluşturur.
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
