package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderFooter_IncludesHelpHints(t *testing.T) {
	out := RenderFooter(NewStyles(), []string{"up/down: navigate", "q: quit"})

	assert.Contains(t, out, "up/down")
	assert.Contains(t, out, "navigate")
	assert.Contains(t, out, "q")
	assert.Contains(t, out, "quit")
}

func TestRenderFooter_DefaultHint(t *testing.T) {
	out := RenderFooter(NewStyles(), nil)

	assert.Contains(t, out, "q")
	assert.Contains(t, out, "quit")
}
