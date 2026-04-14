package email

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultMailgunBaseURL = "https://api.mailgun.net/v3"

type MailgunConfig struct {
	APIKey  string
	Domain  string
	BaseURL string
	Timeout time.Duration
}

func (c MailgunConfig) normalizedBaseURL() string {
	baseURL := strings.TrimSpace(c.BaseURL)
	if baseURL == "" {
		return defaultMailgunBaseURL
	}
	return strings.TrimRight(baseURL, "/")
}

func (c MailgunConfig) normalizedTimeout() time.Duration {
	if c.Timeout <= 0 {
		return 10 * time.Second
	}
	return c.Timeout
}

func (c MailgunConfig) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return fmt.Errorf("mailgun api key is required")
	}
	if strings.TrimSpace(c.Domain) == "" {
		return fmt.Errorf("mailgun domain is required")
	}
	return nil
}

type MailgunSender struct {
	config MailgunConfig
	client *http.Client
}

func NewMailgunSender(config MailgunConfig) (*MailgunSender, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid mailgun config: %w", err)
	}
	return &MailgunSender{
		config: config,
		client: &http.Client{Timeout: config.normalizedTimeout()},
	}, nil
}

func (s *MailgunSender) Send(ctx context.Context, message Message) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("invalid email message: %w", err)
	}

	form := url.Values{}
	form.Set("from", message.From.HeaderValue())

	toValues := make([]string, 0, len(message.To))
	for i := range message.To {
		toValues = append(toValues, message.To[i].HeaderValue())
	}
	form.Set("to", strings.Join(toValues, ","))
	form.Set("subject", message.Subject)

	if strings.HasPrefix(normalizeContentType(message.ContentType), "text/html") {
		form.Set("html", message.Body)
	} else {
		form.Set("text", message.Body)
	}

	endpoint := fmt.Sprintf("%s/%s/messages", s.config.normalizedBaseURL(), s.config.Domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build mailgun request: %w", err)
	}
	req.SetBasicAuth("api", s.config.APIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send mailgun request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("mailgun request failed with status %d", resp.StatusCode)
	}

	return nil
}
