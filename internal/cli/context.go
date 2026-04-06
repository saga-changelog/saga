package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/saga-changelog/saga/internal/config"
	"github.com/saga-changelog/saga/internal/git"
)

// globalDir is set before command execution if --dir is provided.
var globalDir string

// SetGlobalDir is called from main to pass the --dir value.
func SetGlobalDir(dir string) {
	globalDir = dir
}

// sagaDir finds the .saga directory, using the global override if set,
// otherwise by locating the git repo root.
func sagaDir() (string, error) {
	if globalDir != "" {
		if _, err := os.Stat(globalDir); os.IsNotExist(err) {
			return "", fmt.Errorf("saga directory does not exist: %s", globalDir)
		}
		return globalDir, nil
	}

	root, err := git.RepoRoot(".")
	if err != nil {
		return "", err
	}

	dir := filepath.Join(root, ".saga")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("saga is not initialized (no .saga/ directory in %s)", root)
	}

	return dir, nil
}

// loadConfig finds the saga directory and loads the config.
func loadConfig() (string, *config.Config, error) {
	dir, err := sagaDir()
	if err != nil {
		return "", nil, err
	}

	cfg, err := config.Load(dir)
	if err != nil {
		return "", nil, err
	}

	return dir, cfg, nil
}
