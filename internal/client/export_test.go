package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportConferenceBookingsCSV_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conferences/socrates-26/export", r.URL.Path)
		assert.Equal(t, "true", r.URL.Query().Get("include_cancelled"))
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/csv")
		_, _ = w.Write([]byte("name,email\nAlice,alice@test.dev\n"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	data, err := c.ExportConferenceBookingsCSV(context.Background(), "valid-access-token", "socrates-26", true)
	require.NoError(t, err)
	assert.Equal(t, "name,email\nAlice,alice@test.dev\n", string(data))
}

func TestExportConferenceBookingsCSV_MapsErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantErr: ErrUnauthorized},
		{name: "forbidden", statusCode: http.StatusForbidden, wantErr: ErrOrganizerForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantErr: ErrConferenceNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte(`{"error":{"code":"x","message":"x"}}`))
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			data, err := c.ExportConferenceBookingsCSV(context.Background(), "token", "slug", false)
			require.Error(t, err)
			assert.Nil(t, data)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
