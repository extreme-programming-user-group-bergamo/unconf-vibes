package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type testAuthService struct {
	startFn    func(ctx context.Context) (*auth.DeviceAuthorization, error)
	exchangeFn func(ctx context.Context, deviceCode string) (*service.AuthResult, error)
	refreshFn  func(ctx context.Context, refreshToken string) (*service.AuthResult, error)
}

func (s *testAuthService) StartDeviceFlow(ctx context.Context) (*auth.DeviceAuthorization, error) {
	if s.startFn == nil {
		return nil, nil
	}

	return s.startFn(ctx)
}

func (s *testAuthService) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*service.AuthResult, error) {
	if s.exchangeFn == nil {
		return nil, nil
	}

	return s.exchangeFn(ctx, deviceCode)
}

func (s *testAuthService) Refresh(ctx context.Context, refreshToken string) (*service.AuthResult, error) {
	if s.refreshFn == nil {
		return nil, nil
	}

	return s.refreshFn(ctx, refreshToken)
}

func TestAuthHandlerStartDeviceFlowSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&testAuthService{startFn: func(ctx context.Context) (*auth.DeviceAuthorization, error) {
		return &auth.DeviceAuthorization{DeviceCode: "dev", UserCode: "usr", VerificationURI: "https://github.com/login/device", ExpiresIn: 900, Interval: 5}, nil
	}})

	router := gin.New()
	router.POST("/auth/device", handler.StartDeviceFlow)

	req := httptest.NewRequest(http.MethodPost, "/auth/device", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"device_code":"dev"`)
}

func TestAuthHandlerExchangeDeviceCodePending(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&testAuthService{exchangeFn: func(ctx context.Context, deviceCode string) (*service.AuthResult, error) {
		return nil, &service.PendingAuthError{Interval: 10, Cause: service.ErrAuthorizationPending}
	}})

	router := gin.New()
	router.POST("/auth/token", handler.ExchangeDeviceCode)

	payload, _ := json.Marshal(map[string]string{"device_code": "abc"})
	req := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"authorization_pending"`)
	assert.Contains(t, w.Body.String(), `"interval":10`)
}

func TestAuthHandlerRefreshMapsInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&testAuthService{refreshFn: func(ctx context.Context, refreshToken string) (*service.AuthResult, error) {
		return nil, fmt.Errorf("failed to refresh token: %w", service.ErrInvalidRefreshToken)
	}})

	router := gin.New()
	router.POST("/auth/refresh", handler.Refresh)

	payload, _ := json.Marshal(map[string]string{"refresh_token": "bad"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthHandlerExchangeDeviceCodeRateLimitedOnRapidPolling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewAuthHandler(&testAuthService{exchangeFn: func(ctx context.Context, deviceCode string) (*service.AuthResult, error) {
		return nil, &service.PendingAuthError{Interval: 5, Cause: service.ErrAuthorizationPending}
	}})

	router := gin.New()
	router.POST("/auth/token", handler.ExchangeDeviceCode)

	payload, _ := json.Marshal(map[string]string{"device_code": "abc"})
	req1 := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(payload))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusAccepted, w1.Code)

	payload2, _ := json.Marshal(map[string]string{"device_code": "abc"})
	req2 := httptest.NewRequest(http.MethodPost, "/auth/token", bytes.NewReader(payload2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	assert.Contains(t, w2.Body.String(), `"code":"rate_limited"`)
}
