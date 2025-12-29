// Package ui provides shared styling and UI components for subnetCalc.
package ui

import "github.com/charmbracelet/lipgloss"

// Shared lipgloss styles for table rendering across formatter and TUI packages.
var (
	// HeaderStyle is used for table headers.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("39")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	// SelectedStyle highlights the selected row in the TUI.
	SelectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("57")).
			Foreground(lipgloss.Color("15")).
			Bold(true)

	// NormalStyle is the default style for table rows.
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// BorderStyle defines the table border appearance.
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39"))

	// TitleStyle is used for section titles.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			MarginTop(1)

	// StatusStyle displays status messages.
	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82")).
			Bold(true)
)

// PrefixColors is a color palette for different prefix lengths in the TUI.
var PrefixColors = []lipgloss.Color{
	lipgloss.Color("212"), // Pink
	lipgloss.Color("141"), // Purple
	lipgloss.Color("75"),  // Light blue
	lipgloss.Color("81"),  // Cyan
	lipgloss.Color("120"), // Light green
	lipgloss.Color("228"), // Yellow
	lipgloss.Color("216"), // Orange
	lipgloss.Color("210"), // Salmon
	lipgloss.Color("177"), // Light purple
	lipgloss.Color("69"),  // Blue
	lipgloss.Color("87"),  // Light cyan
	lipgloss.Color("156"), // Pale green
	lipgloss.Color("222"), // Gold
	lipgloss.Color("213"), // Light pink
	lipgloss.Color("105"), // Blue-purple
	lipgloss.Color("192"), // Yellow-green
}

// GetColorForPrefix returns a color based on the prefix length.
// It cycles through PrefixColors based on the depth from the initial prefix.
func GetColorForPrefix(bits, initialPrefix int) lipgloss.Color {
	idx := (bits - initialPrefix) % len(PrefixColors)
	if idx < 0 {
		idx = 0
	}
	return PrefixColors[idx]
}
