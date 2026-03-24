package common

import "github.com/charmbracelet/lipgloss"

// Styles groups reusable TUI style definitions.
type Styles struct {
	Header   lipgloss.Style
	Subtle   lipgloss.Style
	Error    lipgloss.Style
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Footer   lipgloss.Style
	HelpKey  lipgloss.Style
}

// NewStyles returns a constrained palette that remains legible on limited-color terminals.
func NewStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}),
		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}),
		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "1", Dark: "9"}),
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}).
			Background(lipgloss.AdaptiveColor{Light: "6", Dark: "4"}),
		Normal: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"}),
		Footer: lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "8", Dark: "8"}),
		HelpKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "4", Dark: "6"}),
	}
}
