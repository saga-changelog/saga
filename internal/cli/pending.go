package cli

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"

	"github.com/saga-changelog/saga/internal/feat"
	"github.com/saga-changelog/saga/internal/theme"
)

// PendingCmd lists pending feats.
type PendingCmd struct{}

func (cmd *PendingCmd) Run() error {
	dir, err := sagaDir()
	if err != nil {
		return err
	}

	pendingDir := filepath.Join(dir, "feats", "pending")
	feats, err := feat.LoadDir(dir, pendingDir)
	if err != nil {
		return fmt.Errorf("loading pending feats: %w", err)
	}

	if len(feats) == 0 {
		fmt.Println(theme.Faint.Render("No pending feats."))
		return nil
	}

	slugs := slices.Sorted(maps.Keys(feats))

	fmt.Println(theme.Emphasis.Render("Pending feats:"))
	fmt.Println()
	for _, s := range slugs {
		f := feats[s]

		// Pad audience names within this feat so titles line up.
		maxAudienceLen := 0
		for _, t := range f.Tales {
			if len(t.Audience) > maxAudienceLen {
				maxAudienceLen = len(t.Audience)
			}
		}

		fmt.Printf("  %s\n", theme.Emphasis.Render(s))
		for _, t := range f.Tales {
			fmt.Printf("    %-*s  %q\n", maxAudienceLen, t.Audience, t.Title)
		}
		fmt.Println()
	}
	fmt.Println(theme.Faint.Render(fmt.Sprintf("%d feats pending", len(feats))))

	return nil
}
