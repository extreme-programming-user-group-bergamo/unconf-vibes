package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutCmd_NotLoggedIn(t *testing.T) {
	store := auth.NewMockTokenStore()
	apiClient := client.NewClient("http://unused")

	cmd := newLogoutCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "You are not currently logged in.")
}

func TestLogoutCmd_SuccessfulLogout(t *testing.T) {
	revokeCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/revoke" {
			revokeCalled = true
			assert.Equal(t, "Bearer access-token-123", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	store.SetTokens("access-token-123", "refresh-token-456")

	cmd := newLogoutCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.True(t, revokeCalled)
	assert.Contains(t, out.String(), "You have been logged out.")
	assert.False(t, store.HasValidToken())
}

func TestLogoutCmd_RevokeFailsGracefully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": "Server error",
			},
		})
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	store.SetTokens("access-token-123", "refresh-token-456")

	cmd := newLogoutCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, out.String(), "You have been logged out.")
	assert.False(t, store.HasValidToken())
}

func TestLogoutCmd_RegisteredInRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	logoutCmd, _, err := cmd.Find([]string{"logout"})
	require.NoError(t, err)
	require.NotNil(t, logoutCmd)
	assert.Equal(t, "logout", logoutCmd.Name())
}
