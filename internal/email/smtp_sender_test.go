package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPSender_ValidatesConfig(t *testing.T) {
	_, err := NewSMTPSender(SMTPConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "smtp host is required")
}

func TestMessageValidate_RequiresFields(t *testing.T) {
	err := Message{}.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid from address")
}
