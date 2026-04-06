package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/saga-changelog/saga/internal/git"
	"github.com/saga-changelog/saga/internal/theme"
)

//go:embed init-agent-instructions.md
var agentInstructions string

// InitCmd initializes Saga in the current git repository.
type InitCmd struct{}

func (cmd *InitCmd) Run() error {
	root, err := git.RepoRoot(".")
	if err != nil {
		return fmt.Errorf("saga init: %w", err)
	}

	sagaDir := filepath.Join(root, ".saga")
	if _, err := os.Stat(sagaDir); err == nil {
		return fmt.Errorf("saga is already initialized in %s", root)
	}

	dirs := []string{
		filepath.Join(sagaDir, "feats", "pending"),
		filepath.Join(sagaDir, "feats", "chapters"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	configPath := filepath.Join(sagaDir, "config.jsonnet")
	if err := os.WriteFile(configPath, []byte(skeletonConfig), 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Println(theme.Success.Render("Saga initialized in " + root))
	fmt.Println()

	// Offer to append saga instructions to agent markdown files.
	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err != nil {
			continue
		}

		var confirm bool
		err := huh.NewConfirm().
			Title(fmt.Sprintf("Add saga instructions to %s?", name)).
			Value(&confirm).
			Run()
		if err != nil || !confirm {
			continue
		}

		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("opening %s: %w", name, err)
		}
		if _, err := f.WriteString(agentInstructions); err != nil {
			f.Close()
			return fmt.Errorf("writing to %s: %w", name, err)
		}
		f.Close()
		fmt.Printf("  Added saga instructions to %s\n", name)
	}

	fmt.Println()
	fmt.Println(theme.Emphasis.Render("Next steps:"))
	fmt.Println("  1. Edit .saga/config.jsonnet to define your audiences and routes")
	fmt.Println("  2. Create feat files in .saga/feats/pending/")
	fmt.Println("  3. Run 'saga validate' to check your configuration")

	return nil
}

const skeletonConfig = `{
  audiences: [
    // {
    //   name: "engineering",
    //   description: "Internal engineering team.",
    //   interest: "API changes, schema migrations, breaking changes.",
    //   tone: "Technical. Reference APIs, schemas, and endpoints.",
    //   routes: [
    //     {
    //       name: "slack-legacy",
    //       courier: {
    //         name: "slack-legacy",
    //         config: {
    //           channel: "#engineering",
    //         },
    //       },
    //     },
    //   ],
    // },
    // {
    //   name: "company",
    //   description: "Company-wide announcements for non-technical stakeholders.",
    //   interest: "New features, workflow changes, important fixes, UX improvements.",
    //   tone: "Friendly and clear. Accessible to non-technical readers. Focus on impact.",
    //   routes: [
    //     {
    //       name: "basecamp-messageboard",
    //       courier: {
    //         name: "basecamp-messageboard",
    //         config: {
    //           project_id: "",
    //           message_board_id: "",
    //         },
    //       },
    //     },
    //   ],
    // },
  ],
}
`
