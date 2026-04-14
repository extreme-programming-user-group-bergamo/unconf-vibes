package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockOrganizerService struct {
	addFn    func(ctx context.Context, slug string, requesterUserID int64, targetUserID int64, role string) (*models.ConferenceOrganizer, error)
	removeFn func(ctx context.Context, slug string, requesterUserID int64, targetUserID int64) error
}

func (m *mockOrganizerService) AddOrganizer(ctx context.Context, slug string, requesterUserID int64, targetUserID int64, role string) (*models.ConferenceOrganizer, error) {
	if m.addFn != nil {
		return m.addFn(ctx, slug, requesterUserID, targetUserID, role)
	}
	return nil, nil
}

func (m *mockOrganizerService) RemoveOrganizer(ctx context.Context, slug string, requesterUserID int64, targetUserID int64) error {
	if m.removeFn != nil {
		return m.removeFn(ctx, slug, requesterUserID, targetUserID)
	}
	return nil
}

func setupOrganizerRouter(handler *OrganizerHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Next()
	})
	router.POST("/conferences/:slug/organizers", handler.Add)
	router.DELETE("/conferences/:slug/organizers/:userID", handler.Remove)
	return router
}

func TestOrganizerHandler_Add_Success(t *testing.T) {
	handler := NewOrganizerHandler(&mockOrganizerService{
		addFn: func(_ context.Context, slug string, requesterUserID int64, targetUserID int64, role string) (*models.ConferenceOrganizer, error) {
			assert.Equal(t, "conf", slug)
			assert.Equal(t, int64(1), requesterUserID)
			assert.Equal(t, int64(2), targetUserID)
			assert.Equal(t, "admin", role)
			return &models.ConferenceOrganizer{ID: 9, ConferenceID: 10, UserID: targetUserID, Role: models.ConferenceOrganizerRoleAdmin}, nil
		},
	})

	router := setupOrganizerRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences/conf/organizers", bytes.NewBufferString(`{"user_id":2,"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"role":"admin"`)
}

func TestOrganizerHandler_Add_Forbidden(t *testing.T) {
	handler := NewOrganizerHandler(&mockOrganizerService{
		addFn: func(_ context.Context, _ string, _ int64, _ int64, _ string) (*models.ConferenceOrganizer, error) {
			return nil, service.ErrOwnerRequired
		},
	})

	router := setupOrganizerRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences/conf/organizers", bytes.NewBufferString(`{"user_id":2,"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestOrganizerHandler_Remove_Success(t *testing.T) {
	handler := NewOrganizerHandler(&mockOrganizerService{
		removeFn: func(_ context.Context, slug string, requesterUserID int64, targetUserID int64) error {
			assert.Equal(t, "conf", slug)
			assert.Equal(t, int64(1), requesterUserID)
			assert.Equal(t, int64(3), targetUserID)
			return nil
		},
	})

	router := setupOrganizerRouter(handler)
	req := httptest.NewRequest(http.MethodDelete, "/conferences/conf/organizers/3", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
