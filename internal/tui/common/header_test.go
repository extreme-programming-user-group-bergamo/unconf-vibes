package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderHeader_IncludesConferenceContext(t *testing.T) {
	out := RenderHeader(NewStyles(), "socrates-26")

	assert.Contains(t, out, "UNCONF Room Explorer")
	assert.Contains(t, out, "Conference: socrates-26")
}

func TestRenderHeader_UsesFallbackContext(t *testing.T) {
	out := RenderHeader(NewStyles(), "")

	assert.Contains(t, out, "Conference: none selected")
}
