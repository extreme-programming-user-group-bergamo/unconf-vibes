package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginCmd_AlreadyLoggedIn(t *testing.T) {
	store := auth.NewMockTokenStore()
	store.SetTokens("existing-token", "existing-refresh")

	apiClient := client.NewClient("http://unused")
	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "You are already logged in")
}

func TestLoginCmd_SuccessfulFlow(t *testing.T) {
	callCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device":
			json.NewEncoder(w).Encode(client.DeviceFlowResponse{
				DeviceCode:      "test-device-code",
				UserCode:        "TEST-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
		case "/auth/token":
			callCount++
			if callCount < 2 {
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(client.PendingResponse{
					Status:   "authorization_pending",
					Interval: 1,
				})
				return
			}
			json.NewEncoder(w).Encode(client.TokenResponse{
				AccessToken:  "access-token-123",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "refresh-token-456",
				User: client.UserResponse{
					ID:          1,
					GitHubID:    "12345",
					Email:       "user@test.com",
					DisplayName: "TestUser",
				},
			})
		}
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "TEST-1234")
	assert.Contains(t, output, "https://github.com/login/device")
	assert.Contains(t, output, "Waiting for authorization...")
	assert.Contains(t, output, "Welcome, TestUser! You are now logged in.")

	accessToken, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "access-token-123", accessToken)

	refreshToken, err := store.GetRefreshToken()
	require.NoError(t, err)
	assert.Equal(t, "refresh-token-456", refreshToken)
}

func TestLoginCmd_ForceRelogin(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device":
			json.NewEncoder(w).Encode(client.DeviceFlowResponse{
				DeviceCode:      "device-code",
				UserCode:        "FORC-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
		case "/auth/token":
			json.NewEncoder(w).Encode(client.TokenResponse{
				AccessToken:  "new-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "new-refresh-token",
				User: client.UserResponse{
					ID:          1,
					GitHubID:    "12345",
					Email:       "user@test.com",
					DisplayName: "ReloggedUser",
				},
			})
		}
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	store.SetTokens("old-token", "old-refresh")

	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"--force"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, out.String(), "Welcome, ReloggedUser!")

	accessToken, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", accessToken)
}

func TestLoginCmd_DeviceFlowServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "service_unavailable",
				"message": "Auth service not configured",
			},
		})
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start login")
}

func TestLoginCmd_AccessDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device":
			json.NewEncoder(w).Encode(client.DeviceFlowResponse{
				DeviceCode:      "device-code",
				UserCode:        "DENY-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
		case "/auth/token":
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "access_denied",
					"message": "Authorization was denied",
				},
			})
		}
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStore()
	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization was denied")
}

func TestLoginCmd_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device":
			json.NewEncoder(w).Encode(client.DeviceFlowResponse{
				DeviceCode:      "device-code",
				UserCode:        "CANC-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
		case "/auth/token":
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(client.PendingResponse{
				Status:   "authorization_pending",
				Interval: 1,
			})
		}
	}))
	defer srv.Close()

	apiClient := client.NewClient(srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately to test pre-cancelled context

	store := auth.NewMockTokenStore()
	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestLoginCmd_RegisteredInRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	loginCmd, _, err := cmd.Find([]string{"login"})
	require.NoError(t, err)
	require.NotNil(t, loginCmd)
	assert.Equal(t, "login", loginCmd.Name())
}

func TestLoginCmd_SaveTokensError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device":
			json.NewEncoder(w).Encode(client.DeviceFlowResponse{
				DeviceCode:      "device-code",
				UserCode:        "SAVE-1234",
				VerificationURI: "https://github.com/login/device",
				ExpiresIn:       900,
				Interval:        1,
			})
		case "/auth/token":
			json.NewEncoder(w).Encode(client.TokenResponse{
				AccessToken:  "token-123",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "refresh-456",
				User: client.UserResponse{
					ID:          1,
					GitHubID:    "12345",
					Email:       "user@test.com",
					DisplayName: "SaveErrUser",
				},
			})
		}
	}))
	defer srv.Close()

	storeErr := errors.New("keyring write failed")
	apiClient := client.NewClient(srv.URL)
	store := auth.NewMockTokenStoreWithError(storeErr)

	cmd := newLoginCmd(apiClient, store)
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save authentication tokens")
	assert.ErrorIs(t, err, storeErr)
}
