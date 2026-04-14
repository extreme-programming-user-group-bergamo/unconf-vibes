package email

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendGridSender_Send(t *testing.T) {
	var authHeader string
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &payload))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	sender, err := NewSendGridSender(SendGridConfig{
		APIKey:  "sg-test-key",
		BaseURL: srv.URL,
	})
	require.NoError(t, err)

	err = sender.Send(context.Background(), Message{
		From:    Address{Email: "from@example.com"},
		To:      []Address{{Email: "to@example.com"}},
		Subject: "Subject",
		Body:    "Body",
		BCC: []Address{
			{Email: "owner@example.com"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Bearer sg-test-key", authHeader)
	personalizations := payload["personalizations"].([]any)
	first := personalizations[0].(map[string]any)
	assert.NotNil(t, first["bcc"])
}
