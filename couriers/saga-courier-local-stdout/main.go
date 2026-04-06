package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saga-changelog/saga/pkg/courier"
)

type stdoutCourier struct{}

func (stdoutCourier) Info() courier.Info {
	return courier.Info{
		Name:        "local-stdout",
		Description: "Prints tales to stdout for testing and development",
		ConfigKeys: []courier.ConfigKey{
			{
				Name:        "prefix",
				Description: "Optional prefix for each output line",
			},
		},
	}
}

func (stdoutCourier) ValidateRoute(_ context.Context, _ *courier.Route) error {
	if os.Getenv("SAGA_COURIER_LOCAL_STDOUT__NOTE") == "" {
		return fmt.Errorf("SAGA_COURIER_LOCAL_STDOUT__NOTE is not set")
	}
	return nil
}

func (stdoutCourier) Tell(_ context.Context, p *courier.Payload) error {
	note := os.Getenv("SAGA_COURIER_LOCAL_STDOUT__NOTE")
	prefix := p.Route.Config["prefix"]

	line := func(s string) {
		if prefix != "" {
			fmt.Printf("[%s] %s\n", prefix, s)
		} else {
			fmt.Println(s)
		}
	}

	line(fmt.Sprintf("=== %s | Route: %s | Audience: %s (note: %s) ===",
		p.Tale.Title, p.Route.Name, p.Route.Audience, note))
	line("")

	for i, b := range p.Tale.Text.Blocks {
		if i > 0 {
			line("")
		}
		text := renderBlock(b)
		for part := range strings.SplitSeq(text, "\n") {
			line(part)
		}
	}

	line("")
	line(fmt.Sprintf("Released in %s", p.Chapter))
	return nil
}

// renderBlock renders a single block to plain text with inline markers
// that mirror the source markdown (**bold**, _italic_, [text](url)).
func renderBlock(b courier.Block) string {
	var sb strings.Builder
	if b.Kind == courier.BlockHeading {
		sb.WriteString("## ")
	}
	for _, in := range b.Inlines {
		sb.WriteString(renderInline(in))
	}
	return sb.String()
}

func renderInline(in courier.Inline) string {
	switch in.Style {
	case courier.InlineBold:
		return "**" + in.Text + "**"
	case courier.InlineItalic:
		return "_" + in.Text + "_"
	case courier.InlineLink:
		return "[" + in.Text + "](" + in.URL + ")"
	default:
		return in.Text
	}
}

func main() {
	courier.Run(stdoutCourier{})
}
