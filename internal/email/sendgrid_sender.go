package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultSendGridBaseURL = "https://api.sendgrid.com/v3/mail/send"

type SendGridConfig struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

func (c SendGridConfig) normalizedBaseURL() string {
	baseURL := strings.TrimSpace(c.BaseURL)
	if baseURL == "" {
		return defaultSendGridBaseURL
	}
	return baseURL
}

func (c SendGridConfig) normalizedTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return c.Timeout
}

func (c SendGridConfig) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("sendgrid api key is required")
	}
	return nil
}

type SendGridSender struct {
	config SendGridConfig
	client *http.Client
}

func NewSendGridSender(config SendGridConfig) (*SendGridSender, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid sendgrid config: %w", err)
	}

	return &SendGridSender{
		config: config,
		client: &http.Client{Timeout: config.normalizedTimeout()},
	}, nil
}

func (s *SendGridSender) Send(ctx context.Context, message Message) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid email message: %w", err)
	}

	to := make([]map[string]string, 0, len(message.To))
	for i := range message.To {
		to = append(to, map[string]string{
			"email": message.To[i].Email,
			"name":  message.To[i].Name,
		})
	}

	payload := map[string]any{
		"personalizations": []any{
			map[string]any{
				"to":      to,
				"subject": message.Subject,
			},
		},
		"from": map[string]string{
			"email": message.From.Email,
			"name":  message.From.Name,
		},
		"content": []any{
			map[string]string{
				"type":  normalizeContentType(message.ContentType),
				"value": message.Body,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal sendgrid payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.normalizedBaseURL(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to build sendgrid request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send sendgrid request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("sendgrid request failed with status %d", resp.StatusCode)
	}

	return nil
}

func normalizeContentType(contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		return "text/plain"
	}
	if strings.HasPrefix(contentType, "text/html") {
		return "text/html"
	}
	return "text/plain"
}
