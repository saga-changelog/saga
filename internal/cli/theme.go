package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/saga-changelog/saga/internal/theme"
)

// ThemeTestCmd displays themed output for visual validation.
type ThemeTestCmd struct {
	Profile string `help:"Force color profile: truecolor, ansi256, ansi, ascii." default:""`
}

func (cmd *ThemeTestCmd) Run() error {
	w, profileDesc := cmd.writer()

	hdr := lipgloss.NewStyle().Bold(true).Foreground(theme.Gold.Color)
	div := lipgloss.NewStyle().Foreground(theme.Rock.Color)
	lbl := lipgloss.NewStyle().Foreground(theme.Rock.Color)

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n", hdr.Render("SAGA THEME TEST"))
	fmt.Fprintf(w, "  %s\n", profileDesc)
	fmt.Fprintln(w)

	// Palette
	fmt.Fprintf(w, "  %s\n\n", div.Render("── Palette "+strings.Repeat("─", 40)))

	block := lipgloss.NewStyle()
	bold := lipgloss.NewStyle().Bold(true)
	plain := lipgloss.NewStyle()

	for _, nc := range theme.Palette() {
		b := block.Foreground(nc.Color)
		n := bold.Foreground(nc.Color)
		d := plain.Foreground(nc.Color)
		fmt.Fprintf(w, "    %s %-10s %s\n", b.Render("██"), n.Render(nc.Name), d.Render(nc.Desc))
	}

	// Semantic styles
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n\n", div.Render("── Styles "+strings.Repeat("─", 41)))

	type semantic struct {
		name  string
		style lipgloss.Style
		text  string
	}
	styles := []semantic{
		{"action", theme.Action, "Going to tell about chapter v1.2.0"},
		{"success", theme.Success, "The saga has been told"},
		{"error", theme.Error, "The scroll could not be read"},
		{"warning", theme.Warning, "The runes fade with time"},
		{"emphasis", theme.Emphasis, "A courier arrives bearing scrolls"},
		{"faint", theme.Faint, "Whispers of forgotten tales"},
	}

	for _, s := range styles {
		fmt.Fprintf(w, "    %s  %s\n", lbl.Render(fmt.Sprintf("%-9s", s.name)), s.style.Render(s.text))
	}

	// Sample output mimicking real saga commands
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n\n", div.Render("── Sample "+strings.Repeat("─", 41)))

	// Init
	fmt.Fprintf(w, "    %s\n", theme.Success.Render("Saga initialized in /home/skald/project"))
	fmt.Fprintf(w, "    %s\n", theme.Emphasis.Render("Next steps:"))
	fmt.Fprintf(w, "      1. Edit .saga/config.jsonnet\n\n")

	// Pending
	fmt.Fprintf(w, "    %s\n\n", theme.Emphasis.Render("Pending feats:"))
	fmt.Fprintf(w, "      %s\n", theme.Emphasis.Render("add-rune-support"))
	fmt.Fprintf(w, "        elders   %q\n", "Add runic inscription API")
	fmt.Fprintf(w, "        seekers  %q\n\n", "New: Rune mode")
	fmt.Fprintf(w, "    %s\n\n", theme.Faint.Render("1 feat pending"))

	// Tell
	fmt.Fprintf(w, "    %s\n", theme.Action.Render("Validating routes..."))
	fmt.Fprintf(w, "    %s elders/raven [add-rune-support]\n", theme.Action.Render("Telling"))

	cname := theme.CourierName.Render("raven")
	sep := theme.StdoutSep.Render(" | ")
	esep := theme.StderrSep.Render(" ! ")
	fmt.Fprintf(w, "      %s%s The tale has been delivered to the elders\n", cname, sep)
	fmt.Fprintf(w, "      %s%s Warning: the northern route grows dark\n", cname, esep)
	fmt.Fprintf(w, "    %s\n\n", theme.Success.Render("1 tales dispatched"))

	// Couriers
	fmt.Fprintf(w, "    %s\n\n", theme.Emphasis.Render("Couriers:"))
	fmt.Fprintf(w, "      raven\n")
	fmt.Fprintf(w, "        %s at /usr/local/bin/saga-courier-raven\n\n", theme.Success.Render("installed"))
	fmt.Fprintf(w, "      scroll-bearer\n")
	fmt.Fprintf(w, "        %s (required by: elders, seers)\n\n", theme.Error.Render("not installed"))
	fmt.Fprintf(w, "      smoke-signal\n")
	fmt.Fprintf(w, "        %s at /usr/local/bin/saga-courier-smoke-signal\n", theme.Warning.Render("info failed"))
	fmt.Fprintln(w)

	return nil
}

var profileMap = map[string]colorprofile.Profile{
	"truecolor": colorprofile.TrueColor,
	"ansi256":   colorprofile.ANSI256,
	"ansi":      colorprofile.ANSI,
	"ascii":     colorprofile.ASCII,
}

var profileDescs = map[colorprofile.Profile]string{
	colorprofile.TrueColor: "24-bit \u00b7 16.7M colors",
	colorprofile.ANSI256:   "8-bit \u00b7 256 colors",
	colorprofile.ANSI:      "4-bit \u00b7 16 colors",
	colorprofile.ASCII:     "no colors",
	colorprofile.NoTTY:     "no terminal",
}

func (cmd *ThemeTestCmd) writer() (io.Writer, string) {
	if cmd.Profile != "" {
		p, ok := profileMap[cmd.Profile]
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown profile %q, using auto-detect\n", cmd.Profile)
		} else {
			desc := fmt.Sprintf("Profile: %s (%s, forced)", p, profileDescs[p])
			return &colorprofile.Writer{Forward: os.Stdout, Profile: p}, desc
		}
	}

	w := colorprofile.NewWriter(os.Stdout, os.Environ())
	desc := fmt.Sprintf("Profile: %s (%s, detected)", w.Profile, profileDescs[w.Profile])
	return w, desc
}
