package email

import (
	"fmt"
	"strings"
)

const (
	ProviderSMTP     = "smtp"
	ProviderSendGrid = "sendgrid"
	ProviderMailgun  = "mailgun"
)

type ProviderConfig struct {
	Provider string
	SMTP     SMTPConfig
	SendGrid SendGridConfig
	Mailgun  MailgunConfig
}

func NewSender(config ProviderConfig) (Sender, error) {
	switch strings.ToLower(strings.TrimSpace(config.Provider)) {
	case ProviderSMTP, "":
		sender, err := NewSMTPSender(config.SMTP)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize smtp sender: %w", err)
		}
		return sender, nil
	case ProviderSendGrid:
		sender, err := NewSendGridSender(config.SendGrid)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize sendgrid sender: %w", err)
		}
		return sender, nil
	case ProviderMailgun:
		sender, err := NewMailgunSender(config.Mailgun)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize mailgun sender: %w", err)
		}
		return sender, nil
	default:
		return nil, fmt.Errorf("unsupported email provider %q", config.Provider)
	}
}
