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
		_ = json.NewEncoder(w).Encode(expected)
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
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "unauthorized",
						"message": "Token expired",
					},
				})
				return
			}
			// Second call (after refresh): succeed
			assert.Equal(t, "Bearer new-access-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(UserResponse{
				ID:          42,
				GitHubID:    "12345",
				Email:       "user@example.com",
				DisplayName: "octocat",
			})

		case "/auth/refresh":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			assert.Equal(t, "valid-refresh-token", body["refresh_token"])

			_ = json.NewEncoder(w).Encode(TokenResponse{
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "Token expired",
				},
			})

		case "/auth/refresh":
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
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
		_ = json.NewEncoder(w).Encode(map[string]any{
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
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "Token expired",
				},
			})

		case "/auth/refresh":
			// Refresh succeeds
			_ = json.NewEncoder(w).Encode(TokenResponse{
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

func TestAuthenticatedClient_UpdateMe_Success(t *testing.T) {
	displayName := "Updated"
	input := UpdateProfileRequest{DisplayName: &displayName}

	expected := UserResponse{
		ID:             42,
		GitHubID:       "12345",
		Email:          "user@example.com",
		DisplayName:    "Updated",
		PrivacySetting: "public",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.UpdateMe(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "Updated", user.DisplayName)
	assert.Equal(t, "public", user.PrivacySetting)
}

func TestAuthenticatedClient_UpdateMe_AutoRefresh(t *testing.T) {
	var callCount atomic.Int32
	displayName := "Updated"
	input := UpdateProfileRequest{DisplayName: &displayName}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			n := callCount.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "unauthorized",
						"message": "Token expired",
					},
				})
				return
			}
			assert.Equal(t, "Bearer new-access-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(UserResponse{
				ID:             42,
				GitHubID:       "12345",
				Email:          "user@example.com",
				DisplayName:    "Updated",
				PrivacySetting: "public",
			})

		case "/auth/refresh":
			_ = json.NewEncoder(w).Encode(TokenResponse{
				AccessToken:  "new-access-token",
				TokenType:    "Bearer",
				ExpiresIn:    86400,
				RefreshToken: "new-refresh-token",
				User: UserResponse{
					ID:          42,
					GitHubID:    "12345",
					Email:       "user@example.com",
					DisplayName: "Updated",
				},
			})
		}
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("expired-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	user, err := ac.UpdateMe(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, "Updated", user.DisplayName)

	newAccess, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", newAccess)
}

func TestAuthenticatedClient_UpdateMe_RefreshFails_ReturnsSessionExpired(t *testing.T) {
	displayName := "Updated"
	input := UpdateProfileRequest{DisplayName: &displayName}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/users/me":
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"code":    "unauthorized",
					"message": "Token expired",
				},
			})

		case "/auth/refresh":
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{
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
	user, err := ac.UpdateMe(context.Background(), input)

	assert.Nil(t, user)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.False(t, store.HasValidToken())
}

func TestAuthenticatedClient_CreateBooking_Success(t *testing.T) {
	input := CreateBookingRequest{RoomID: 10, ConferenceID: 2, PrivacySetting: "public", Notes: "vegan meal"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bookings", r.URL.Path)
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BookingResponse{ID: 77, RoomID: 10, ConferenceID: 2, Status: "requested", PrivacySetting: "public"})
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	booking, err := ac.CreateBooking(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, int64(77), booking.ID)
	assert.Equal(t, "requested", booking.Status)
}

func TestAuthenticatedClient_CreateBooking_AutoRefresh(t *testing.T) {
	var bookingsCallCount atomic.Int32
	input := CreateBookingRequest{RoomID: 10, ConferenceID: 2, PrivacySetting: "public"}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/bookings":
			n := bookingsCallCount.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "unauthorized",
						"message": "token expired",
					},
				})
				return
			}

			assert.Equal(t, "Bearer new-access-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode(BookingResponse{ID: 99, RoomID: 10, ConferenceID: 2, Status: "requested", PrivacySetting: "public"})
		case "/auth/refresh":
			_ = json.NewEncoder(w).Encode(TokenResponse{
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
	booking, err := ac.CreateBooking(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, int64(99), booking.ID)
	assert.Equal(t, int32(2), bookingsCallCount.Load())

	newAccess, err := store.GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "new-access-token", newAccess)
}

func TestAuthenticatedClient_ListBookings_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/bookings", r.URL.Path)
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]BookingResponse{{ID: 101, ConferenceID: 2, Status: "confirmed"}})
	}))
	defer srv.Close()

	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	ac := NewAuthenticatedClient(NewClient(srv.URL), store)
	bookings, err := ac.ListBookings(context.Background())

	require.NoError(t, err)
	require.Len(t, bookings, 1)
	assert.Equal(t, int64(101), bookings[0].ID)
}

func TestAuthenticatedClient_ListRoommateRequests_AutoRefresh(t *testing.T) {
	var requestCallCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/requests":
			n := requestCallCount.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "unauthorized",
						"message": "token expired",
					},
				})
				return
			}

			assert.Equal(t, "Bearer new-access-token", r.Header.Get("Authorization"))
			_ = json.NewEncoder(w).Encode([]RoommateRequestResponse{{ID: 1, Status: "pending", Direction: "incoming"}})
		case "/auth/refresh":
			_ = json.NewEncoder(w).Encode(TokenResponse{
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
	requests, err := ac.ListRoommateRequests(context.Background())

	require.NoError(t, err)
	require.Len(t, requests, 1)
	assert.Equal(t, int32(2), requestCallCount.Load())
}
