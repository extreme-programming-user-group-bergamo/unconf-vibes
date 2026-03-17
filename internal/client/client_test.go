package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartDeviceFlow_Success(t *testing.T) {
	expected := DeviceFlowResponse{
		DeviceCode:      "device-123",
		UserCode:        "ABCD-1234",
		VerificationURI: "https://github.com/login/device",
		ExpiresIn:       900,
		Interval:        5,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/device", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.StartDeviceFlow(context.Background())

	require.NoError(t, err)
	assert.Equal(t, expected.DeviceCode, resp.DeviceCode)
	assert.Equal(t, expected.UserCode, resp.UserCode)
	assert.Equal(t, expected.VerificationURI, resp.VerificationURI)
	assert.Equal(t, expected.ExpiresIn, resp.ExpiresIn)
	assert.Equal(t, expected.Interval, resp.Interval)
}

func TestStartDeviceFlow_ServerError(t *testing.T) {
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

	c := NewClient(srv.URL)
	resp, err := c.StartDeviceFlow(context.Background())

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start device flow")
}

func TestExchangeDeviceCode_Success(t *testing.T) {
	expected := TokenResponse{
		AccessToken:  "paseto-token",
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		RefreshToken: "refresh-token",
		User: UserResponse{
			ID:          1,
			GitHubID:    "12345",
			Email:       "user@example.com",
			DisplayName: "octocat",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/token", r.URL.Path)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "device-123", body["device_code"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "device-123")

	require.NoError(t, err)
	assert.Equal(t, expected.AccessToken, resp.AccessToken)
	assert.Equal(t, expected.User.DisplayName, resp.User.DisplayName)
}

func TestExchangeDeviceCode_Pending(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(PendingResponse{
			Status:   "authorization_pending",
			Interval: 5,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "device-123")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrAuthorizationPending)
}

func TestExchangeDeviceCode_SlowDown(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(PendingResponse{
			Status:   "slow_down",
			Interval: 10,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "device-123")

	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrSlowDown)
}

func TestExchangeDeviceCode_AccessDenied(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "access_denied",
				"message": "Authorization was denied",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "device-123")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestExchangeDeviceCode_ExpiredToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "expired_token",
				"message": "Device authorization has expired",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "device-123")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrExpiredDeviceCode)
}

func TestExchangeDeviceCode_InvalidDeviceCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "invalid_device_code",
				"message": "Device code is invalid",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(context.Background(), "bad-code")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Device code is invalid")
}

func TestRefreshToken_Success(t *testing.T) {
	expected := TokenResponse{
		AccessToken:  "new-access-token",
		TokenType:    "Bearer",
		ExpiresIn:    86400,
		RefreshToken: "new-refresh-token",
		User: UserResponse{
			ID:          1,
			GitHubID:    "12345",
			Email:       "user@example.com",
			DisplayName: "octocat",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/refresh", r.URL.Path)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "old-refresh-token", body["refresh_token"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.RefreshToken(context.Background(), "old-refresh-token")

	require.NoError(t, err)
	assert.Equal(t, "new-access-token", resp.AccessToken)
	assert.Equal(t, "new-refresh-token", resp.RefreshToken)
}

func TestRefreshToken_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "invalid_refresh_token",
				"message": "Refresh token is invalid",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.RefreshToken(context.Background(), "bad-token")

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to refresh token")
}

func TestRevokeToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/auth/revoke", r.URL.Path)
		assert.Equal(t, "Bearer my-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.RevokeToken(context.Background(), "my-token")

	assert.NoError(t, err)
}

func TestRevokeToken_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": "Something went wrong",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	err := c.RevokeToken(context.Background(), "my-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke token")
}

func TestExchangeDeviceCode_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient(srv.URL)
	resp, err := c.ExchangeDeviceCode(ctx, "device-123")

	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestGetMe_Success(t *testing.T) {
	expected := UserResponse{
		ID:          42,
		GitHubID:    "12345",
		Email:       "user@example.com",
		DisplayName: "octocat",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/users/me", r.URL.Path)
		assert.Equal(t, "Bearer my-access-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	user, err := c.GetMe(context.Background(), "my-access-token")

	require.NoError(t, err)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "octocat", user.DisplayName)
	assert.Equal(t, "user@example.com", user.Email)
}

func TestGetMe_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "Invalid or expired token",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	user, err := c.GetMe(context.Background(), "bad-token")

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestGetMe_NetworkError(t *testing.T) {
	c := NewClient("http://127.0.0.1:1") // connection refused
	user, err := c.GetMe(context.Background(), "token")

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user profile")
}

func TestListConferences_Success(t *testing.T) {
	expected := []ConferenceResponse{
		{
			ID:            1,
			Slug:          "gophercon-2026",
			Name:          "GopherCon 2026",
			Description:   "Go conference",
			Location:      "Denver, CO",
			StartDate:     "2026-06-15",
			EndDate:       "2026-06-18",
			Capacity:      500,
			AttendeeCount: 120,
			Status:        "upcoming",
		},
		{
			ID:            2,
			Slug:          "rustconf-2026",
			Name:          "RustConf 2026",
			Description:   "Rust conference",
			Location:      "Portland, OR",
			StartDate:     "2026-08-01",
			EndDate:       "2026-08-03",
			Capacity:      300,
			AttendeeCount: 50,
			Status:        "upcoming",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/conferences", r.URL.Path)
		assert.Empty(t, r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expected)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	conferences, err := c.ListConferences(context.Background())

	require.NoError(t, err)
	require.Len(t, conferences, 2)
	assert.Equal(t, "GopherCon 2026", conferences[0].Name)
	assert.Equal(t, "RustConf 2026", conferences[1].Name)
	assert.Equal(t, 120, conferences[0].AttendeeCount)
}

func TestListConferences_EmptyArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	conferences, err := c.ListConferences(context.Background())

	require.NoError(t, err)
	require.NotNil(t, conferences)
	assert.Empty(t, conferences)
}

func TestListConferences_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"code":    "internal_error",
				"message": "database unavailable",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	conferences, err := c.ListConferences(context.Background())

	assert.Nil(t, conferences)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list conferences")
	assert.Contains(t, err.Error(), "500")
}

func TestListConferences_NetworkError(t *testing.T) {
	c := NewClient("http://127.0.0.1:1") // connection refused
	conferences, err := c.ListConferences(context.Background())

	assert.Nil(t, conferences)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list conferences")
}
