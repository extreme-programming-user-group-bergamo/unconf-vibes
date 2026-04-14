package email

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailgunSender_Send(t *testing.T) {
	var authUser string
	var authPass string
	var formValues url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authUser, authPass, _ = r.BasicAuth()
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		formValues, err = url.ParseQuery(string(body))
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	sender, err := NewMailgunSender(MailgunConfig{
		APIKey:  "mg-test-key",
		Domain:  "mg.example.com",
		BaseURL: srv.URL,
	})
	require.NoError(t, err)

	err = sender.Send(context.Background(), Message{
		From: Address{Email: "from@example.com"},
		To: []Address{
			{Email: "to@example.com", Name: "Primary Recipient"},
			{Email: "john@example.com", Name: "Doe, John"},
		},
		Subject: "Subject",
		Body:    "Body",
	})
	require.NoError(t, err)
	assert.Equal(t, "api", authUser)
	assert.Equal(t, "mg-test-key", authPass)
	assert.Contains(t, formValues.Get("to"), "\"Primary Recipient\" <to@example.com>")
	assert.Contains(t, formValues.Get("to"), "\"Doe, John\" <john@example.com>")
	assert.NotContains(t, formValues.Get("to"), "\r")
	assert.NotContains(t, formValues.Get("to"), "\n")
}
