package feat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/saga-changelog/saga/internal/safejsonnet"
)

// Feat represents a single change record. A feat holds one or more
// audience-specific tales; it carries no audience-agnostic description
// of its own, because saga's whole point is that each audience gets its
// own wording. The feat's identity comes from its slug (filename).
type Feat struct {
	Tales []Tale `json:"tales"`
}

// Tale is an audience-specific description of a change. Title is a
// plain-text, single-line heading used by couriers as the message
// subject/header. Text is a markdown string (parsed via taletext) that
// couriers render into their native format.
type Tale struct {
	Audience string `json:"audience"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}

// Slug returns the feat's identifier derived from its filename.
func Slug(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), ".jsonnet")
}

// LoadFile reads and evaluates a single feat jsonnet file.
// sagaDir is the .saga/ directory used to sandbox jsonnet imports.
func LoadFile(sagaDir, path string) (*Feat, error) {
	vm, err := safejsonnet.MakeVM(sagaDir)
	if err != nil {
		return nil, fmt.Errorf("creating jsonnet VM: %w", err)
	}
	return loadFileWithVM(vm, path)
}

// loadFileWithVM evaluates a single feat file using an existing VM.
func loadFileWithVM(vm *jsonnet.VM, path string) (*Feat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading feat: %w", err)
	}

	jsonStr, err := vm.EvaluateAnonymousSnippet(path, string(data))
	if err != nil {
		return nil, fmt.Errorf("evaluating feat jsonnet: %w", err)
	}

	var f Feat
	if err := json.Unmarshal([]byte(jsonStr), &f); err != nil {
		return nil, fmt.Errorf("deserializing feat: %w", err)
	}

	return &f, nil
}

// LoadDir reads and evaluates all .jsonnet files in a directory.
// sagaDir is the .saga/ directory used to sandbox jsonnet imports.
// It returns a map of slug to feat.
func LoadDir(sagaDir, dir string) (map[string]*Feat, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	vm, err := safejsonnet.MakeVM(sagaDir)
	if err != nil {
		return nil, fmt.Errorf("creating jsonnet VM: %w", err)
	}

	feats := make(map[string]*Feat)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonnet") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := loadFileWithVM(vm, path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		feats[Slug(e.Name())] = f
	}

	return feats, nil
}
