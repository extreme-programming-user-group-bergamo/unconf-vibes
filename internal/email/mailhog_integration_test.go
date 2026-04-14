package email

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMailHog_NewBookingRenderAndSend(t *testing.T) {
	if strings.TrimSpace(os.Getenv("UNCONF_MAILHOG_TEST")) != "1" {
		t.Skip("set UNCONF_MAILHOG_TEST=1 to run MailHog integration test")
	}

	smtpPort, err := strconv.Atoi(getEnvOrDefault("UNCONF_SMTP_PORT", "1025"))
	require.NoError(t, err)

	renderer := NewTemplateRenderer()
	sender, err := NewSMTPSender(SMTPConfig{
		Host: getEnvOrDefault("UNCONF_SMTP_HOST", "localhost"),
		Port: smtpPort,
	})
	require.NoError(t, err)

	service, err := NewService(renderer, sender, Address{
		Email: getEnvOrDefault("UNCONF_EMAIL_FROM_ADDRESS", "noreply@unconf.local"),
		Name:  "UNCONF",
	})
	require.NoError(t, err)

	err = service.SendBookingEmail(
		context.Background(),
		TemplateTypeNewBooking,
		[]Address{{Email: getEnvOrDefault("UNCONF_MAILHOG_TO", "hotel@example.test")}},
		TemplateData{
			GuestName:       "MailHog Test Guest",
			RoomNumber:      "901",
			StartDate:       "2026-05-10",
			EndDate:         "2026-05-12",
			SpecialRequests: "Late checkout",
		},
	)
	require.NoError(t, err)

	subject, body, err := waitForMailHogMessage(context.Background(), getEnvOrDefault("UNCONF_MAILHOG_API_URL", "http://localhost:8025"))
	require.NoError(t, err)
	require.Contains(t, subject, "New Booking")
	require.Contains(t, body, "MailHog Test Guest")
	require.Contains(t, body, "901")
	require.Contains(t, body, "Late checkout")
}

func waitForMailHogMessage(ctx context.Context, apiBaseURL string) (string, string, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	url := strings.TrimRight(apiBaseURL, "/") + "/api/v2/messages"

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", "", fmt.Errorf("failed to build mailhog request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			subject, body, parseErr := parseMailHogResponse(resp)
			if parseErr == nil {
				return subject, body, nil
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	return "", "", fmt.Errorf("timed out waiting for mailhog message")
}

func parseMailHogResponse(resp *http.Response) (string, string, error) {
	type mailhogItem struct {
		Content struct {
			Body    string              `json:"Body"`
			Headers map[string][]string `json:"Headers"`
		} `json:"Content"`
	}
	type mailhogResponse struct {
		Items []mailhogItem `json:"items"`
	}

	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("mailhog API returned status %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read mailhog response body: %w", err)
	}

	var parsed mailhogResponse
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return "", "", fmt.Errorf("failed to decode mailhog response: %w", err)
	}
	if len(parsed.Items) == 0 {
		return "", "", fmt.Errorf("no messages available")
	}

	item := parsed.Items[len(parsed.Items)-1]
	subject := ""
	if values, ok := item.Content.Headers["Subject"]; ok && len(values) > 0 {
		subject = values[0]
	}
	return subject, item.Content.Body, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
