package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockConferenceService struct {
	listConferencesFn  func(ctx context.Context) ([]*service.ConferenceResponse, error)
	getConferenceFn    func(ctx context.Context, slug string) (*service.ConferenceResponse, error)
	createConferenceFn func(ctx context.Context, creatorUserID int64, input service.CreateConferenceInput) (*service.ConferenceResponse, error)
	updateConferenceFn func(ctx context.Context, slug string, input service.UpdateConferenceInput) (*service.ConferenceResponse, error)
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

func (m *mockConferenceService) CreateConference(ctx context.Context, creatorUserID int64, input service.CreateConferenceInput) (*service.ConferenceResponse, error) {
	if m.createConferenceFn != nil {
		return m.createConferenceFn(ctx, creatorUserID, input)
	}
	return nil, nil
}

func (m *mockConferenceService) UpdateConference(ctx context.Context, slug string, input service.UpdateConferenceInput) (*service.ConferenceResponse, error) {
	if m.updateConferenceFn != nil {
		return m.updateConferenceFn(ctx, slug, input)
	}
	return nil, nil
}

func setupConferenceRouter(handler *ConferenceHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(77))
		c.Next()
	})
	router.GET("/conferences", handler.List)
	router.GET("/conferences/:slug", handler.GetBySlug)
	router.POST("/conferences", handler.Create)
	router.PUT("/conferences/:slug", handler.Update)

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

func TestConferenceHandler_Create_Success(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		createConferenceFn: func(_ context.Context, creatorUserID int64, input service.CreateConferenceInput) (*service.ConferenceResponse, error) {
			assert.Equal(t, int64(77), creatorUserID)
			assert.Equal(t, "new-conf", input.Slug)
			assert.Equal(t, time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC), input.StartDate)
			return &service.ConferenceResponse{
				ID:            42,
				Slug:          input.Slug,
				Name:          input.Name,
				Description:   input.Description,
				Location:      input.Location,
				StartDate:     "2026-10-07",
				EndDate:       "2026-10-10",
				Capacity:      input.Capacity,
				AttendeeCount: 0,
				Status:        "upcoming",
			}, nil
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences", bytes.NewBufferString(`{"slug":"new-conf","name":"New Conf","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"slug":"new-conf"`)
}

func TestConferenceHandler_Create_DuplicateSlug(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		createConferenceFn: func(_ context.Context, _ int64, _ service.CreateConferenceInput) (*service.ConferenceResponse, error) {
			return nil, service.ErrConferenceExists
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences", bytes.NewBufferString(`{"slug":"dup","name":"Dup","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), `"conflict"`)
}

func TestConferenceHandler_Create_OrganizerRequired(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		createConferenceFn: func(_ context.Context, _ int64, _ service.CreateConferenceInput) (*service.ConferenceResponse, error) {
			return nil, service.ErrOrganizerForbidden
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences", bytes.NewBufferString(`{"slug":"new","name":"New","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), `"forbidden"`)
}

func TestConferenceHandler_Update_Success(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		updateConferenceFn: func(_ context.Context, slug string, input service.UpdateConferenceInput) (*service.ConferenceResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, "Updated", input.Name)
			return &service.ConferenceResponse{
				ID:            1,
				Slug:          slug,
				Name:          input.Name,
				Description:   input.Description,
				Location:      input.Location,
				StartDate:     "2026-10-07",
				EndDate:       "2026-10-10",
				Capacity:      input.Capacity,
				AttendeeCount: 0,
				Status:        "upcoming",
			}, nil
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodPut, "/conferences/socrates-26", bytes.NewBufferString(`{"name":"Updated","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":150}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"slug":"socrates-26"`)
	assert.Contains(t, w.Body.String(), `"name":"Updated"`)
}

func TestConferenceHandler_Update_NotFound(t *testing.T) {
	handler := NewConferenceHandler(&mockConferenceService{
		updateConferenceFn: func(_ context.Context, _ string, _ service.UpdateConferenceInput) (*service.ConferenceResponse, error) {
			return nil, service.ErrConferenceNotFound
		},
	})

	router := setupConferenceRouter(handler)
	req := httptest.NewRequest(http.MethodPut, "/conferences/missing", bytes.NewBufferString(`{"name":"Updated","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":150}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), `"not_found"`)
}
