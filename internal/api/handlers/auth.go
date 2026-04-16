package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/service"
)

type AuthHandler struct {
	authService  serviceAuthService
	pollThrottle *pollThrottle
}

const defaultTokenPollMinInterval = time.Second

type pollThrottle struct {
	mu          sync.Mutex
	nextAllowed map[string]time.Time
	minInterval time.Duration
	nowFunc     func() time.Time
}

type serviceAuthService interface {
	StartDeviceFlow(ctx context.Context) (*auth.DeviceAuthorization, error)
	ExchangeDeviceCode(ctx context.Context, deviceCode string) (*service.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (*service.AuthResult, error)
	RevokeSession(ctx context.Context, sessionID int64) error
	ListActiveSessions(ctx context.Context, userID int64, currentSessionID int64) ([]service.SessionView, error)
	RevokeSessionForUser(ctx context.Context, userID int64, sessionID int64, currentSessionID int64) error
	RevokeOtherSessions(ctx context.Context, userID int64, currentSessionID int64) (int64, error)
}

type exchangeTokenRequest struct {
	DeviceCode string `json:"device_code"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(authService serviceAuthService) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		pollThrottle: newPollThrottle(defaultTokenPollMinInterval),
	}
}

func newPollThrottle(minInterval time.Duration) *pollThrottle {
	if minInterval <= 0 {
		minInterval = defaultTokenPollMinInterval
	}

	return &pollThrottle{
		nextAllowed: make(map[string]time.Time),
		minInterval: minInterval,
		nowFunc:     time.Now,
	}
}

func (t *pollThrottle) allow(clientKey string) (bool, int) {
	if t == nil {
		return true, 0
	}

	now := t.nowFunc().UTC()
	if clientKey == "" {
		return true, 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if next, ok := t.nextAllowed[clientKey]; ok && now.Before(next) {
		retryAfter := int(next.Sub(now).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}

		return false, retryAfter
	}

	t.nextAllowed[clientKey] = now.Add(t.minInterval)
	return true, 0
}

func (h *AuthHandler) StartDeviceFlow(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	result, err := h.authService.StartDeviceFlow(c.Request.Context())
	if err != nil {
		slog.Error("failed to start device flow", "error", err)
		responses.WriteError(c, "auth_device_failed", "Failed to start device authorization flow", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) ExchangeDeviceCode(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	if allowed, retryAfter := h.pollThrottle.allow(c.ClientIP()); !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"code":      "rate_limited",
				"message":   "Too many token polling requests",
				"timestamp": time.Now().UTC(),
			},
			"retry_after_seconds": retryAfter,
		})
		return
	}

	req := exchangeTokenRequest{}
	if err := c.ShouldBindJSON(&req); err != nil || req.DeviceCode == "" {
		responses.WriteError(c, "invalid_request", "device_code is required", http.StatusBadRequest)
		return
	}

	ctxWithMetadata := service.WithSessionClientMetadata(c.Request.Context(), buildClientMetadata(c))
	result, err := h.authService.ExchangeDeviceCode(ctxWithMetadata, req.DeviceCode)
	if err != nil {
		if pending := mapPendingError(c, err); pending {
			return
		}

		status, code, message := mapAuthError(err)
		slog.Error("failed to exchange device code", "error", err, "status", status, "code", code)
		responses.WriteError(c, code, message, status)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	req := refreshTokenRequest{}
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		responses.WriteError(c, "invalid_request", "refresh_token is required", http.StatusBadRequest)
		return
	}

	ctxWithMetadata := service.WithSessionClientMetadata(c.Request.Context(), buildClientMetadata(c))
	result, err := h.authService.Refresh(ctxWithMetadata, req.RefreshToken)
	if err != nil {
		status, code, message := mapAuthError(err)
		slog.Error("failed to refresh token", "error", err, "status", status, "code", code)
		responses.WriteError(c, code, message, status)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AuthHandler) Revoke(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	sid, ok := middleware.GetSessionID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "Session ID not found in token", http.StatusUnauthorized)
		return
	}

	if err := h.authService.RevokeSession(c.Request.Context(), sid); err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			responses.WriteError(c, "session_not_found", "Session not found or already revoked", http.StatusUnauthorized)
			return
		}

		slog.Error("failed to revoke session", "error", err)
		responses.WriteError(c, "internal_error", "Failed to revoke session", http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in token", http.StatusUnauthorized)
		return
	}

	currentSessionID, ok := middleware.GetSessionID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "Session ID not found in token", http.StatusUnauthorized)
		return
	}

	sessions, err := h.authService.ListActiveSessions(c.Request.Context(), userID, currentSessionID)
	if err != nil {
		slog.Error("failed to list active sessions", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to list active sessions", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func (h *AuthHandler) RevokeSessionByID(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in token", http.StatusUnauthorized)
		return
	}

	currentSessionID, ok := middleware.GetSessionID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "Session ID not found in token", http.StatusUnauthorized)
		return
	}

	sessionID, err := strconv.ParseInt(c.Param("session_id"), 10, 64)
	if err != nil || sessionID <= 0 {
		responses.WriteError(c, "invalid_request", "session_id must be a positive integer", http.StatusBadRequest)
		return
	}

	if err := h.authService.RevokeSessionForUser(c.Request.Context(), userID, sessionID, currentSessionID); err != nil {
		switch {
		case errors.Is(err, service.ErrCannotRevokeCurrentSession):
			responses.WriteError(c, "invalid_request", "Cannot revoke current session with this endpoint", http.StatusBadRequest)
		case errors.Is(err, service.ErrSessionNotFound):
			responses.WriteError(c, "session_not_found", "Session not found or already revoked", http.StatusNotFound)
		default:
			slog.Error("failed to revoke selected session", "error", err, "user_id", userID, "session_id", sessionID)
			responses.WriteError(c, "internal_error", "Failed to revoke selected session", http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) RevokeOthers(c *gin.Context) {
	if h.authService == nil {
		responses.WriteError(c, "service_unavailable", "Authentication service is not configured", http.StatusServiceUnavailable)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in token", http.StatusUnauthorized)
		return
	}

	currentSessionID, ok := middleware.GetSessionID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "Session ID not found in token", http.StatusUnauthorized)
		return
	}

	revokedCount, err := h.authService.RevokeOtherSessions(c.Request.Context(), userID, currentSessionID)
	if err != nil {
		slog.Error("failed to revoke other sessions", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to revoke other sessions", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"revoked_count": revokedCount})
}

func buildClientMetadata(c *gin.Context) string {
	parts := make([]string, 0, 2)

	userAgent := strings.TrimSpace(c.GetHeader("User-Agent"))
	if userAgent != "" {
		parts = append(parts, "ua="+userAgent)
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	if clientIP != "" {
		parts = append(parts, "ip="+clientIP)
	}

	metadata := strings.Join(parts, ";")
	if len(metadata) > 255 {
		return metadata[:255]
	}

	return metadata
}

func mapPendingError(c *gin.Context, err error) bool {
	var pendingErr *service.PendingAuthError
	if !errors.As(err, &pendingErr) {
		return false
	}

	status := "authorization_pending"
	if errors.Is(err, service.ErrSlowDown) {
		status = "slow_down"
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":   status,
		"interval": pendingErr.Interval,
	})

	return true
}

func mapAuthError(err error) (int, string, string) {
	switch {
	case errors.Is(err, service.ErrInvalidDeviceCode):
		return http.StatusBadRequest, "invalid_device_code", "Device code is invalid"
	case errors.Is(err, service.ErrAccessDenied):
		return http.StatusUnauthorized, "access_denied", "Authorization was denied"
	case errors.Is(err, service.ErrExpiredDeviceCode):
		return http.StatusUnauthorized, "expired_token", "Device authorization has expired"
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return http.StatusUnauthorized, "invalid_refresh_token", "Refresh token is invalid"
	case errors.Is(err, service.ErrExpiredRefreshToken):
		return http.StatusUnauthorized, "expired_refresh_token", "Refresh token has expired"
	case errors.Is(err, service.ErrRevokedRefreshToken):
		return http.StatusUnauthorized, "revoked_refresh_token", "Refresh token has been revoked"
	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error"
	}
}
