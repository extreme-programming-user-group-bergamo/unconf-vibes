package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/auth"
)

const (
	contextKeyUserID    = "user_id"
	contextKeySessionID = "session_id"
)

// TokenValidator defines the interface for validating PASETO tokens.
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*auth.AccessTokenClaims, error)
}

// AuthMiddleware returns a gin.HandlerFunc that validates Bearer tokens.
func AuthMiddleware(validator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			responses.WriteError(c, "unauthorized", "Authorization header is required", http.StatusUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(header, "Bearer ") {
			responses.WriteError(c, "unauthorized", "Authorization header must use Bearer scheme", http.StatusUnauthorized)
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == "" {
			responses.WriteError(c, "unauthorized", "Bearer token is empty", http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, err := validator.ValidateToken(c.Request.Context(), tokenStr)
		if err != nil {
			responses.WriteError(c, "unauthorized", "Invalid or expired token", http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeySessionID, claims.SessionID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from the gin context.
func GetUserID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(contextKeyUserID)
	if !exists {
		return 0, false
	}

	userID, ok := val.(int64)
	return userID, ok
}

// GetSessionID retrieves the authenticated session ID from the gin context.
func GetSessionID(c *gin.Context) (int64, bool) {
	val, exists := c.Get(contextKeySessionID)
	if !exists {
		return 0, false
	}

	sessionID, ok := val.(int64)
	return sessionID, ok
}
