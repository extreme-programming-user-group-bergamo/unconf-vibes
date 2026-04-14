package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockOrganizerPermissionChecker struct {
	isOrganizerFn func(ctx context.Context, slug string, userID int64) (bool, error)
	isOwnerFn     func(ctx context.Context, slug string, userID int64) (bool, error)
}

func (m *mockOrganizerPermissionChecker) IsOrganizerForConference(ctx context.Context, slug string, userID int64) (bool, error) {
	if m.isOrganizerFn != nil {
		return m.isOrganizerFn(ctx, slug, userID)
	}
	return false, nil
}

func (m *mockOrganizerPermissionChecker) IsOwnerForConference(ctx context.Context, slug string, userID int64) (bool, error) {
	if m.isOwnerFn != nil {
		return m.isOwnerFn(ctx, slug, userID)
	}
	return false, nil
}

func TestRequireOrganizer_ForbiddenForNonOrganizer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(10))
		c.Next()
	})
	router.GET("/conferences/:slug/attendees", RequireOrganizer(&mockOrganizerPermissionChecker{
		isOrganizerFn: func(_ context.Context, _ string, _ int64) (bool, error) {
			return false, nil
		},
	}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/conferences/test-conf/attendees", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestRequireOrganizer_AllowsOrganizer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(11))
		c.Next()
	})
	router.GET("/conferences/:slug/attendees", RequireOrganizer(&mockOrganizerPermissionChecker{
		isOrganizerFn: func(_ context.Context, _ string, _ int64) (bool, error) {
			return true, nil
		},
	}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/conferences/test-conf/attendees", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireOrganizerOwner_ConferenceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(12))
		c.Next()
	})
	router.POST("/conferences/:slug/organizers", RequireOrganizerOwner(&mockOrganizerPermissionChecker{
		isOwnerFn: func(_ context.Context, _ string, _ int64) (bool, error) {
			return false, service.ErrConferenceNotFound
		},
	}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/conferences/missing/organizers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRequireOrganizerOwner_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(12))
		c.Next()
	})
	router.POST("/conferences/:slug/organizers", RequireOrganizerOwner(&mockOrganizerPermissionChecker{
		isOwnerFn: func(_ context.Context, _ string, _ int64) (bool, error) {
			return false, errors.New("db down")
		},
	}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/conferences/missing/organizers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
