package theme

import (
	"image/color"

	lipgloss "charm.land/lipgloss/v2"
)

// NamedColor pairs a palette color with its name and description.
type NamedColor struct {
	Name  string
	Color color.Color
	Desc  string
}

// Saga palette — colors drawn from the lore of ancient sagas.
//
// Defined as TrueColor hex values. Lipgloss downsamples automatically
// through colorprofile.Writer for terminals that don't support 24-bit.
var (
	Gold     = NamedColor{"Gold", lipgloss.Color("#DAA520"), "Precious gold, action and triumph"}
	Olive    = NamedColor{"Olive", lipgloss.Color("#808000"), "Greek olive groves, success"}
	BloodRed = NamedColor{"BloodRed", lipgloss.Color("#8B0000"), "Sacrifice, war, error"}
	Ember    = NamedColor{"Ember", lipgloss.Color("#CC5500"), "The fading fire, warning"}
	Marble   = NamedColor{"Marble", lipgloss.Color("#E6DDD1"), "Pale noble columns, important neutral"}
	Rock     = NamedColor{"Rock", lipgloss.Color("#808080"), "Weathered grey stone, deemphasis"}
	Rune     = NamedColor{"Rune", lipgloss.Color("#4A90A4"), "Nordic carved inscriptions, unassigned"}
)

// Palette returns all named colors for display/testing purposes.
func Palette() []NamedColor {
	return []NamedColor{Gold, Olive, BloodRed, Ember, Marble, Rock, Rune}
}

// Semantic styles used across the CLI for consistent theming.
var (
	Action   = lipgloss.NewStyle().Foreground(Gold.Color).Bold(true)
	Success  = lipgloss.NewStyle().Foreground(Olive.Color)
	Warning  = lipgloss.NewStyle().Foreground(Ember.Color)
	Error    = lipgloss.NewStyle().Foreground(BloodRed.Color)
	Emphasis = lipgloss.NewStyle().Foreground(Marble.Color)
	Faint    = lipgloss.NewStyle().Foreground(Rock.Color)

	CourierName = lipgloss.NewStyle().Foreground(Marble.Color).Bold(true)
	StdoutSep   = lipgloss.NewStyle().Foreground(Marble.Color)
	StderrSep   = lipgloss.NewStyle().Foreground(BloodRed.Color).Bold(true)
)
