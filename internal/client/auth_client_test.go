package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatedClient_GetMe_Success(t *testing.T) {
	expected := UserResponse{
		ID:          42,
		GitHubID:    "12345",
		Email:       "user@example.com",
		DisplayName: "octocat",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "octocat", user.DisplayName)
	assert.Equal(t, "user@example.com", user.Email)
}

func TestAuthenticatedClient_GetMe_AutoRefresh(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			n := callCount.Add(1)
			if n == 1 {
				// First call: reject with 401
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "unauthorized",
						"message": "Token expired",
					},
				})
				return
			}
			// Second call (after refresh): succeed
			assert.Equal(t, "Bearer new-access-token", r.Header.Get("Authorization"))
			json.NewEncoder(w).Encode(UserResponse{
				ID:          42,
				GitHubID:    "12345",
				Email:       "user@example.com",
				DisplayName: "octocat",
			})

		case "/auth/refresh":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "valid-refresh-token", body["refresh_token"])

			json.NewEncoder(w).Encode(TokenResponse{
				AccessToken:  "new-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "new-refresh-token",
				User: UserResponse{
					ID:          42,
					GitHubID:    "12345",
					Email:       "user@example.com",
					DisplayName: "octocat",
				},
			})
		}
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("expired-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "octocat", user.DisplayName)

	// Verify tokens were updated in store
	newAccess, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", newAccess)

	newRefresh, err := store.GetRefreshToken()
	require.NoError(t, err)
	assert.Equal(t, "new-refresh-token", newRefresh)
}

func TestAuthenticatedClient_GetMe_RefreshFailure_ReturnsSessionExpired(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "Token expired",
				},
			})

		case "/auth/refresh":
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "invalid_refresh_token",
					"message": "Refresh token is invalid",
				},
			})
		}
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("expired-access-token", "invalid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	assert.Nil(t, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionExpired)

	// Verify tokens were cleared
	assert.False(t, store.HasValidToken())
}

func TestAuthenticatedClient_GetMe_NotAuthenticated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("should not make any HTTP calls when not authenticated")
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	// No tokens set

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	assert.Nil(t, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
}

func TestAuthenticatedClient_GetMe_NoRefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Always return 401 for /users/me
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "Token expired",
			},
		})
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	// Set only access token, no refresh token
	store.SetTokens("expired-access-token", "")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	assert.Nil(t, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionExpired)

	// Verify tokens were cleared
	assert.False(t, store.HasValidToken())
}

func TestAuthenticatedClient_GetMe_NetworkError(t *testing.T) {
	// Use a server that is immediately closed to simulate network error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	assert.Nil(t, user)
	require.Error(t, err)
	// Should NOT be ErrSessionExpired — network errors propagate directly
	assert.NotErrorIs(t, err, ErrSessionExpired)
	// Tokens should still be in store (not cleared on network error)
	assert.True(t, store.HasValidToken())
}

func TestAuthenticatedClient_GetMe_RefreshSucceedsButRetryFails(t *testing.T) {
	var meCallCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			// Both calls return 401
			meCallCount.Add(1)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "Token expired",
				},
			})

		case "/auth/refresh":
			// Refresh succeeds
			json.NewEncoder(w).Encode(TokenResponse{
				AccessToken:  "new-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "new-refresh-token",
			})
		}
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("expired-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.GetMe(context.Background())

	assert.Nil(t, user)
	require.Error(t, err)
	// Should have made exactly 2 GetMe calls (original + retry)
	assert.Equal(t, int32(2), meCallCount.Load())
	// Should NOT infinite loop — just returns the error
	assert.ErrorIs(t, err, ErrUnauthorized)
}
