package theme

import (
	lipgloss "charm.land/lipgloss/v2"
)

// Styles used across the CLI for consistent theming.
var (
	// CourierName styles the courier name in prefixed output.
	CourierName = lipgloss.NewStyle().Foreground(lipgloss.Cyan).Bold(true)

	// StdoutSep styles the ":" separator for courier stdout lines.
	StdoutSep = lipgloss.NewStyle().Foreground(lipgloss.Cyan)

	// StderrSep styles the "!" separator for courier stderr lines.
	StderrSep = lipgloss.NewStyle().Foreground(lipgloss.Red).Bold(true)

	// Success styles success messages.
	Success = lipgloss.NewStyle().Foreground(lipgloss.Green)

	// Error styles error messages.
	Error = lipgloss.NewStyle().Foreground(lipgloss.Red)

	// Warning styles warning messages.
	Warning = lipgloss.NewStyle().Foreground(lipgloss.Yellow)

	// Faint styles de-emphasized text.
	Faint = lipgloss.NewStyle().Faint(true)
)
