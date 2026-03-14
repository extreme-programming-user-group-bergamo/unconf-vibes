package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTokenValidator struct {
	validateFn func(ctx context.Context, token string) (*auth.AccessTokenClaims, error)
}

func (m *mockTokenValidator) ValidateToken(ctx context.Context, token string) (*auth.AccessTokenClaims, error) {
	if m.validateFn != nil {
		return m.validateFn(ctx, token)
	}

	return nil, fmt.Errorf("no validate function set")
}

func setupAuthRouter(validator TokenValidator) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware(validator))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		sessionID, _ := GetSessionID(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"session_id": sessionID,
		})
	})

	return router
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	validator := &mockTokenValidator{
		validateFn: func(_ context.Context, token string) (*auth.AccessTokenClaims, error) {
			assert.Equal(t, "valid-token", token)
			return &auth.AccessTokenClaims{UserID: 42, SessionID: 99}, nil
		},
	}

	router := setupAuthRouter(validator)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":42`)
	assert.Contains(t, w.Body.String(), `"session_id":99`)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	validator := &mockTokenValidator{}
	router := setupAuthRouter(validator)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	validator := &mockTokenValidator{}
	router := setupAuthRouter(validator)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic something")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Bearer scheme")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	validator := &mockTokenValidator{
		validateFn: func(_ context.Context, _ string) (*auth.AccessTokenClaims, error) {
			return nil, fmt.Errorf("invalid token")
		},
	}

	router := setupAuthRouter(validator)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	validator := &mockTokenValidator{}
	router := setupAuthRouter(validator)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	userID, ok := GetUserID(c)

	assert.False(t, ok)
	assert.Equal(t, int64(0), userID)
}

func TestGetSessionID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	sessionID, ok := GetSessionID(c)

	assert.False(t, ok)
	assert.Equal(t, int64(0), sessionID)
}

func TestGetUserID_Set(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(42))

	userID, ok := GetUserID(c)
	require.True(t, ok)
	assert.Equal(t, int64(42), userID)
}
