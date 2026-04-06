package chapter

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/saga-changelog/saga/internal/feat"
)

// Chapter represents a versioned collection of feats.
type Chapter struct {
	Version string
	Feats   map[string]*feat.Feat
}

// List returns all chapters found in the saga chapters directory, sorted by semver.
func List(sagaDir string) ([]Chapter, error) {
	chaptersDir := filepath.Join(sagaDir, "feats", "chapters")
	entries, err := os.ReadDir(chaptersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading chapters directory: %w", err)
	}

	var chapters []Chapter
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		chapters = append(chapters, Chapter{Version: e.Name()})
	}

	sort.Slice(chapters, func(i, j int) bool {
		vi, ei := semver.NewVersion(chapters[i].Version)
		vj, ej := semver.NewVersion(chapters[j].Version)
		if ei != nil || ej != nil {
			// Unparseable versions: fall back to string comparison.
			return chapters[i].Version < chapters[j].Version
		}
		return vi.LessThan(vj)
	})

	return chapters, nil
}

// Load reads all feats for a specific chapter version.
func Load(sagaDir, version string) (*Chapter, error) {
	chapterDir := filepath.Join(sagaDir, "feats", "chapters", version)

	if _, err := os.Stat(chapterDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("chapter %q does not exist", version)
	}

	feats, err := feat.LoadDir(sagaDir, chapterDir)
	if err != nil {
		return nil, fmt.Errorf("loading chapter %s: %w", version, err)
	}

	return &Chapter{
		Version: version,
		Feats:   feats,
	}, nil
}

// Create bundles all pending feats into a new chapter.
func Create(sagaDir, version string) ([]string, error) {
	if version == "pending" {
		return nil, fmt.Errorf("chapter name %q is reserved", version)
	}

	chapterDir := filepath.Join(sagaDir, "feats", "chapters", version)
	pendingDir := filepath.Join(sagaDir, "feats", "pending")

	if _, err := os.Stat(chapterDir); err == nil {
		return nil, fmt.Errorf("chapter %q already exists", version)
	}

	pending, err := feat.LoadDir(sagaDir, pendingDir)
	if err != nil {
		return nil, fmt.Errorf("loading pending feats: %w", err)
	}
	if len(pending) == 0 {
		return nil, fmt.Errorf("no pending feats")
	}

	slugs := slices.Sorted(maps.Keys(pending))

	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating chapter directory: %w", err)
	}

	for _, s := range slugs {
		src := filepath.Join(pendingDir, s+".jsonnet")
		dst := filepath.Join(chapterDir, s+".jsonnet")
		if err := os.Rename(src, dst); err != nil {
			return nil, fmt.Errorf("moving feat %s: %w", s, err)
		}
	}

	return slugs, nil
}
