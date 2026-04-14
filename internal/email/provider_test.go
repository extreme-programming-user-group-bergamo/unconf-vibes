package email

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSender_SelectsProvider(t *testing.T) {
	tests := []struct {
		name     string
		config   ProviderConfig
		wantType string
	}{
		{
			name: "smtp provider",
			config: ProviderConfig{
				Provider: ProviderSMTP,
				SMTP: SMTPConfig{
					Host: "localhost",
					Port: 1025,
				},
			},
			wantType: "*email.SMTPSender",
		},
		{
			name: "sendgrid provider",
			config: ProviderConfig{
				Provider: ProviderSendGrid,
				SendGrid: SendGridConfig{
					APIKey: "test-key",
				},
			},
			wantType: "*email.SendGridSender",
		},
		{
			name: "mailgun provider",
			config: ProviderConfig{
				Provider: ProviderMailgun,
				Mailgun: MailgunConfig{
					APIKey: "test-key",
					Domain: "mg.example.com",
				},
			},
			wantType: "*email.MailgunSender",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender, err := NewSender(tt.config)
			require.NoError(t, err)
			assert.Equal(t, tt.wantType, fmt.Sprintf("%T", sender))
		})
	}
}

func TestNewSender_RejectsUnknownProvider(t *testing.T) {
	_, err := NewSender(ProviderConfig{Provider: "unknown"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported email provider")
}
