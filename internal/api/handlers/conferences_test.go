package handlers

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

type mockConferenceService struct {
	listConferencesFn func(ctx context.Context) ([]*service.ConferenceResponse, error)
	getConferenceFn   func(ctx context.Context, slug string) (*service.ConferenceResponse, error)
}

func (m *mockConferenceService) ListConferences(ctx context.Context) ([]*service.ConferenceResponse, error) {
	if m.listConferencesFn != nil {
		return m.listConferencesFn(ctx)
	}

	return nil, nil
}

func (m *mockConferenceService) GetConference(ctx context.Context, slug string) (*service.ConferenceResponse, error) {
	if m.getConferenceFn != nil {
		return m.getConferenceFn(ctx, slug)
	}

	return nil, nil
}

func setupConferenceRouter(handler *ConferenceHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/conferences", handler.List)
	router.GET("/conferences/:slug", handler.GetBySlug)

	return router
}

func testConferenceResponse() *service.ConferenceResponse {
	return &service.ConferenceResponse{
		ID:            1,
		Slug:          "socrates-26",
		Name:          "SoCraTes 2026",
		Description:   "Software Craftsmanship and Testing Conference",
		Location:      "Saarbrücken, Germany",
		StartDate:     "2026-10-07",
		EndDate:       "2026-10-10",
		Capacity:      200,
		AttendeeCount: 0,
		Status:        "upcoming",
	}
}

func TestConferenceHandler_List_Success(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		listConferencesFn: func(_ context.Context) ([]*service.ConferenceResponse, error) {
			return []*service.ConferenceResponse{testConferenceResponse()}, nil
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"slug":"socrates-26"`)
	assert.Contains(t, w.Body.String(), `"status":"upcoming"`)
	assert.Contains(t, w.Body.String(), `"attendee_count":0`)
}

func TestConferenceHandler_List_Empty(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		listConferencesFn: func(_ context.Context) ([]*service.ConferenceResponse, error) {
			return []*service.ConferenceResponse{}, nil
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestConferenceHandler_List_ServiceError(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		listConferencesFn: func(_ context.Context) ([]*service.ConferenceResponse, error) {
			return nil, errors.New("db error")
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

func TestConferenceHandler_GetBySlug_Success(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		getConferenceFn: func(_ context.Context, slug string) (*service.ConferenceResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			return testConferenceResponse(), nil
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"slug":"socrates-26"`)
	assert.Contains(t, w.Body.String(), `"status":"upcoming"`)
	assert.Contains(t, w.Body.String(), `"attendee_count":0`)
}

func TestConferenceHandler_GetBySlug_NotFound(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		getConferenceFn: func(_ context.Context, _ string) (*service.ConferenceResponse, error) {
			return nil, service.ErrConferenceNotFound
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
}

func TestConferenceHandler_GetBySlug_ServiceError(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		getConferenceFn: func(_ context.Context, _ string) (*service.ConferenceResponse, error) {
			return nil, errors.New("db error")
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/some-slug", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}
