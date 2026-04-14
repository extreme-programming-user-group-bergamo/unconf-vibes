package email

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendGridSender_Send(t *testing.T) {
	var authHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
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
	})
	require.NoError(t, err)
	assert.Equal(t, "Bearer sg-test-key", authHeader)
}
