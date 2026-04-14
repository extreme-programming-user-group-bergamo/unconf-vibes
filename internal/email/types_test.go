package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressValidate_AcceptsRFCCompatibleEmail(t *testing.T) {
	err := Address{Email: "valid@example.com", Name: "Valid Name"}.Validate()
	require.NoError(t, err)
}

func TestAddressValidate_RejectsCRLFInjection(t *testing.T) {
	tests := []Address{
		{Email: "valid@example.com\r\nBcc:evil@example.com"},
		{Email: "valid@example.com", Name: "Valid\r\nName"},
	}

	for i := range tests {
		err := tests[i].Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	}
}

func TestAddressValidate_RejectsInvalidEmail(t *testing.T) {
	err := Address{Email: "invalid-email", Name: "Name"}.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email address is invalid")
}

func TestAddressHeaderValue_UsesSafeFormatting(t *testing.T) {
	header := Address{Email: "john@example.com", Name: "Doe, John"}.HeaderValue()
	assert.Equal(t, "\"Doe, John\" <john@example.com>", header)
	assert.NotContains(t, header, "\r")
	assert.NotContains(t, header, "\n")
}
