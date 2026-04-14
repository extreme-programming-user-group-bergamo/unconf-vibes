package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	UseTLS   bool
}

func (c SMTPConfig) Address() string {
	return fmt.Sprintf("%s:%d", strings.TrimSpace(c.Host), c.Port)
}

func (c SMTPConfig) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return fmt.Errorf("smtp host is required")
	}
	if c.Port <= 0 {
		return fmt.Errorf("smtp port must be positive")
	}
	if strings.TrimSpace(c.Password) != "" && strings.TrimSpace(c.Username) == "" {
		return fmt.Errorf("smtp username is required when password is set")
	}
	return nil
}

type SMTPSender struct {
	config SMTPConfig
}

func NewSMTPSender(config SMTPConfig) (*SMTPSender, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid smtp config: %w", err)
	}
	return &SMTPSender{config: config}, nil
}

func (s *SMTPSender) Send(ctx context.Context, message Message) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("smtp send canceled: %w", err)
	}
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid email message: %w", err)
	}

	var auth smtp.Auth
	if strings.TrimSpace(s.config.Username) != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, strings.TrimSpace(s.config.Host))
	}

	if s.config.UseTLS {
		return s.sendWithTLS(ctx, auth, message)
	}

	rawMessage, recipientEmails := buildSMTPMessage(message)
	if err := smtp.SendMail(s.config.Address(), auth, message.From.Email, recipientEmails, rawMessage); err != nil {
		return fmt.Errorf("failed to send email via smtp: %w", err)
	}

	return nil
}

func (s *SMTPSender) sendWithTLS(ctx context.Context, auth smtp.Auth, message Message) error {
	dialer := &tls.Dialer{
		Config: &tls.Config{
			ServerName: strings.TrimSpace(s.config.Host),
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", s.config.Address())
	if err != nil {
		return fmt.Errorf("failed to dial smtp server with tls: %w", err)
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, strings.TrimSpace(s.config.Host))
	if err != nil {
		return fmt.Errorf("failed to create smtp client: %w", err)
	}
	defer func() { _ = client.Close() }()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate smtp client: %w", err)
		}
	}

	if err := client.Mail(message.From.Email); err != nil {
		return fmt.Errorf("failed to set smtp sender address: %w", err)
	}
	for i := range message.To {
		if err := client.Rcpt(message.To[i].Email); err != nil {
			return fmt.Errorf("failed to set smtp recipient address: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open smtp data writer: %w", err)
	}

	rawMessage, _ := buildSMTPMessage(message)
	if _, err := writer.Write(rawMessage); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write smtp message body: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close smtp message writer: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close smtp session cleanly: %w", err)
	}

	return nil
}

func buildSMTPMessage(message Message) ([]byte, []string) {
	recipients := make([]string, 0, len(message.To))
	toHeaderValues := make([]string, 0, len(message.To))
	for i := range message.To {
		recipients = append(recipients, message.To[i].Email)
		toHeaderValues = append(toHeaderValues, message.To[i].HeaderValue())
	}

	contentType := strings.TrimSpace(message.ContentType)
	if contentType == "" {
		contentType = DefaultContentType()
	}

	rawMessage := strings.Join([]string{
		fmt.Sprintf("From: %s", message.From.HeaderValue()),
		fmt.Sprintf("To: %s", strings.Join(toHeaderValues, ", ")),
		fmt.Sprintf("Subject: %s", message.Subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: %s", contentType),
		"",
		message.Body,
	}, "\r\n")

	return []byte(rawMessage), recipients
}
