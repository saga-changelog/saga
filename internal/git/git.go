package git

import (
	"fmt"
	"os"
	"path/filepath"
)

// RepoRoot returns the root directory of the git repository containing dir.
// It walks up the directory tree looking for a .git directory.
func RepoRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	for {
		if info, err := os.Stat(filepath.Join(abs, ".git")); err == nil && info.IsDir() {
			return abs, nil
		}

		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("not a git repository (or any parent up to /)")
		}
		abs = parent
	}
}
