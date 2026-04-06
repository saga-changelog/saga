package safejsonnet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
)

const (
	// maxStack limits recursion depth to catch infinite recursion
	// without burning through the host's stack or CPU.
	maxStack = 200
)

// MakeVM returns a jsonnet VM that is safe to run on untrusted input.
//
//   - MaxStack is capped to prevent runaway recursion.
//   - Imports are restricted to files under allowDir (the .saga/ directory),
//     preventing traversal into the rest of the repository or system.
//   - No native functions are registered.
func MakeVM(allowDir string) (*jsonnet.VM, error) {
	abs, err := filepath.Abs(allowDir)
	if err != nil {
		return nil, fmt.Errorf("resolving allow directory: %w", err)
	}

	vm := jsonnet.MakeVM()
	vm.MaxStack = maxStack
	vm.Importer(&sandboxImporter{root: abs})
	return vm, nil
}

// sandboxImporter restricts file imports to paths under root.
type sandboxImporter struct {
	root string
}

func (si *sandboxImporter) Import(importedFrom, importedPath string) (jsonnet.Contents, string, error) {
	// Resolve relative to the importing file's directory.
	var dir string
	if importedFrom == "" {
		dir = si.root
	} else {
		dir = filepath.Dir(importedFrom)
	}

	resolved := filepath.Clean(filepath.Join(dir, importedPath))

	// Ensure the resolved path is under the allowed root.
	if !strings.HasPrefix(resolved, si.root+string(filepath.Separator)) && resolved != si.root {
		return jsonnet.Contents{}, "", fmt.Errorf(
			"import %q resolves to %q which is outside the allowed directory %q",
			importedPath, resolved, si.root)
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return jsonnet.Contents{}, "", err
	}

	return jsonnet.MakeContents(string(data)), resolved, nil
}
