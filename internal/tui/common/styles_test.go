package common

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestNewStyles_LimitedColorPaletteIntent(t *testing.T) {
	styles := NewStyles()

	assert.True(t, styles.Header.GetBold())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "0", Dark: "15"}, styles.Header.GetForeground())

	assert.Equal(t, lipgloss.AdaptiveColor{Light: "8", Dark: "8"}, styles.Subtle.GetForeground())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "8", Dark: "8"}, styles.Footer.GetForeground())

	assert.True(t, styles.Error.GetBold())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "1", Dark: "9"}, styles.Error.GetForeground())

	assert.True(t, styles.Selected.GetBold())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "0", Dark: "15"}, styles.Selected.GetForeground())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "6", Dark: "4"}, styles.Selected.GetBackground())

	assert.Equal(t, lipgloss.AdaptiveColor{Light: "0", Dark: "15"}, styles.Normal.GetForeground())

	assert.True(t, styles.HelpKey.GetBold())
	assert.Equal(t, lipgloss.AdaptiveColor{Light: "4", Dark: "6"}, styles.HelpKey.GetForeground())
}
