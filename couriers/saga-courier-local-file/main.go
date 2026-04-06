package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saga-changelog/saga/pkg/courier"
)

type fileCourier struct{}

func (fileCourier) Info() courier.Info {
	return courier.Info{
		Name:        "local-file",
		Description: "Appends tales to a local file in markdown format",
		ConfigKeys:  []courier.ConfigKey{},
	}
}

func (fileCourier) ValidateRoute(_ context.Context, _ *courier.Route) error {
	path := os.Getenv("SAGA_COURIER_LOCAL_FILE__PATH")
	if path == "" {
		return fmt.Errorf("SAGA_COURIER_LOCAL_FILE__PATH is not set")
	}

	// Verify the parent directory exists so we fail early rather than
	// during tell when a tale has already been partially delivered.
	dir := path[:max(strings.LastIndex(path, "/"), 0)]
	if dir != "" {
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("parent directory of SAGA_COURIER_LOCAL_FILE__PATH: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("parent of SAGA_COURIER_LOCAL_FILE__PATH is not a directory: %s", dir)
		}
	}

	return nil
}

func (fileCourier) Tell(_ context.Context, p *courier.Payload) error {
	path := os.Getenv("SAGA_COURIER_LOCAL_FILE__PATH")
	if path == "" {
		return fmt.Errorf("SAGA_COURIER_LOCAL_FILE__PATH is not set")
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	var b strings.Builder

	b.WriteString("## ")
	b.WriteString(p.Tale.Title)
	b.WriteString("\n")

	text := renderText(p.Tale.Text)
	if text != "" {
		b.WriteString("\n")
		b.WriteString(text)
		b.WriteString("\n")
	}

	b.WriteString("\n_Released in ")
	b.WriteString(p.Chapter)
	b.WriteString("_\n\n")

	if _, err := f.WriteString(b.String()); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	fmt.Printf("Appended to %s\n", path)
	return nil
}

// renderText renders a TaleText to markdown. Blocks are separated by
// blank lines.
func renderText(text courier.TaleText) string {
	var b strings.Builder
	for i, block := range text.Blocks {
		if i > 0 {
			b.WriteString("\n\n")
		}
		switch block.Kind {
		case courier.BlockHeading:
			b.WriteString("### ")
			b.WriteString(renderInlines(block.Inlines))
		default:
			b.WriteString(renderInlines(block.Inlines))
		}
	}
	return b.String()
}

// renderInlines flattens a list of Inlines to a markdown string.
func renderInlines(inlines []courier.Inline) string {
	var b strings.Builder
	for _, in := range inlines {
		switch in.Style {
		case courier.InlineBold:
			b.WriteString("**")
			b.WriteString(in.Text)
			b.WriteString("**")
		case courier.InlineItalic:
			b.WriteString("_")
			b.WriteString(in.Text)
			b.WriteString("_")
		case courier.InlineLink:
			b.WriteString("[")
			b.WriteString(in.Text)
			b.WriteString("](")
			b.WriteString(in.URL)
			b.WriteString(")")
		default:
			b.WriteString(in.Text)
		}
	}
	return b.String()
}

func main() {
	courier.Run(fileCourier{})
}
