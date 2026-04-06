package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/saga-changelog/saga/internal/chapter"
)

// ChaptersCmd lists chapters or manages them.
type ChaptersCmd struct {
	Create ChaptersCreateCmd `cmd:"" help:"Bundle pending feats into a new chapter."`
	List   ChaptersListCmd   `cmd:"" default:"withargs" help:"List all chapters."`
}

// ChaptersListCmd lists all chapters.
type ChaptersListCmd struct{}

func (cmd *ChaptersListCmd) Run() error {
	dir, err := sagaDir()
	if err != nil {
		return err
	}

	chapters, err := chapter.List(dir)
	if err != nil {
		return fmt.Errorf("listing chapters: %w", err)
	}

	if len(chapters) == 0 {
		fmt.Println("No chapters.")
		return nil
	}

	fmt.Println("Chapters:")
	fmt.Println()
	for _, c := range chapters {
		n, err := countFeats(filepath.Join(dir, "feats", "chapters", c.Version))
		if err != nil {
			return err
		}
		fmt.Printf("  %s  (%d feats)\n", c.Version, n)
	}
	fmt.Println()
	fmt.Printf("%d chapters\n", len(chapters))

	return nil
}

// countFeats counts .jsonnet files in a directory without parsing them.
func countFeats(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, fmt.Errorf("reading directory %s: %w", dir, err)
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".jsonnet") {
			n++
		}
	}
	return n, nil
}

// ChaptersCreateCmd bundles all pending feats into a new chapter.
type ChaptersCreateCmd struct {
	Version string `arg:"" help:"Version string for the new chapter."`
}

func (cmd *ChaptersCreateCmd) Run() error {
	dir, _, err := loadConfig()
	if err != nil {
		return err
	}

	moved, err := chapter.Create(dir, cmd.Version)
	if err != nil {
		return fmt.Errorf("creating chapter: %w", err)
	}

	fmt.Printf("Created chapter %s with %d feats:\n", cmd.Version, len(moved))
	for _, s := range moved {
		fmt.Printf("  %s\n", s)
	}

	return nil
}
