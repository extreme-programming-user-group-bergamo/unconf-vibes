package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
)

type mockRoomService struct {
	listRoomsFn  func(ctx context.Context, slug string) ([]*service.RoomResponse, error)
	createRoomFn func(ctx context.Context, slug string, input service.ManageRoomInput) (*models.Room, error)
	updateRoomFn func(ctx context.Context, slug string, roomNumber string, input service.ManageRoomInput) (*models.Room, error)
	deleteRoomFn func(ctx context.Context, slug string, roomNumber string) error
}

func (m *mockRoomService) ListRooms(ctx context.Context, slug string) ([]*service.RoomResponse, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx, slug)
	}
	return nil, nil
}

func (m *mockRoomService) CreateRoom(ctx context.Context, slug string, input service.ManageRoomInput) (*models.Room, error) {
	if m.createRoomFn != nil {
		return m.createRoomFn(ctx, slug, input)
	}
	return nil, nil
}

func (m *mockRoomService) UpdateRoom(ctx context.Context, slug string, roomNumber string, input service.ManageRoomInput) (*models.Room, error) {
	if m.updateRoomFn != nil {
		return m.updateRoomFn(ctx, slug, roomNumber, input)
	}
	return nil, nil
}

func (m *mockRoomService) DeleteRoom(ctx context.Context, slug string, roomNumber string) error {
	if m.deleteRoomFn != nil {
		return m.deleteRoomFn(ctx, slug, roomNumber)
	}
	return nil
}

func setupRoomRouter(handler *RoomHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/conferences/:slug/rooms", handler.ListByConference)
	router.POST("/conferences/:slug/rooms", handler.Create)
	router.PUT("/conferences/:slug/rooms/:number", handler.Update)
	router.DELETE("/conferences/:slug/rooms/:number", handler.Delete)
	return router
}

func testRoomResponse() *service.RoomResponse {
	return &service.RoomResponse{
		ID:             1,
		ConferenceID:   1,
		RoomNumber:     "101",
		RoomType:       "double",
		PricePerNight:  120.0,
		Capacity:       2,
		SpotsTaken:     1,
		SpotsAvailable: 1,
		Occupants: []service.OccupantResponse{
			{DisplayName: "Alice", UserID: int64Ptr(10)},
		},
	}
}

func int64Ptr(v int64) *int64 { return &v }

func TestRoomHandler_ListByConference_Success(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		listRoomsFn: func(_ context.Context, slug string) ([]*service.RoomResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			return []*service.RoomResponse{testRoomResponse()}, nil
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/rooms", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"room_number":"101"`)
	assert.Contains(t, w.Body.String(), `"spots_taken":1`)
	assert.Contains(t, w.Body.String(), `"spots_available":1`)
}

func TestRoomHandler_ListByConference_Empty(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		listRoomsFn: func(_ context.Context, _ string) ([]*service.RoomResponse, error) {
			return []*service.RoomResponse{}, nil
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/rooms", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestRoomHandler_ListByConference_ConferenceNotFound(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		listRoomsFn: func(_ context.Context, _ string) ([]*service.RoomResponse, error) {
			return nil, service.ErrConferenceNotFound
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/nonexistent/rooms", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
}

func TestRoomHandler_ListByConference_ServiceError(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		listRoomsFn: func(_ context.Context, _ string) ([]*service.RoomResponse, error) {
			return nil, errors.New("db error")
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/rooms", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

func TestRoomHandler_Create_Success(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		createRoomFn: func(_ context.Context, slug string, input service.ManageRoomInput) (*models.Room, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, "101", input.RoomNumber)
			return &models.Room{ID: 1, ConferenceID: 2, RoomNumber: "101", RoomType: "double", PricePerNight: 120, Capacity: 2}, nil
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodPost, "/conferences/socrates-26/rooms", bytes.NewBufferString(`{"room_number":"101","room_type":"double","price_per_night":120,"capacity":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"room_number":"101"`)
}

func TestRoomHandler_Update_Success(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		updateRoomFn: func(_ context.Context, slug string, roomNumber string, input service.ManageRoomInput) (*models.Room, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, "101", roomNumber)
			assert.Equal(t, "102", input.RoomNumber)
			return &models.Room{ID: 1, ConferenceID: 2, RoomNumber: "102", RoomType: "triple", PricePerNight: 180, Capacity: 3}, nil
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodPut, "/conferences/socrates-26/rooms/101", bytes.NewBufferString(`{"room_number":"102","room_type":"triple","price_per_night":180,"capacity":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"room_number":"102"`)
}

func TestRoomHandler_Update_RoomNotFound(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		updateRoomFn: func(_ context.Context, _ string, _ string, _ service.ManageRoomInput) (*models.Room, error) {
			return nil, service.ErrRoomNotFound
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodPut, "/conferences/socrates-26/rooms/999", bytes.NewBufferString(`{"room_number":"999","room_type":"double","price_per_night":120,"capacity":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
}

func TestRoomHandler_Delete_HasBookingsConflict(t *testing.T) {
	handler := NewRoomHandler(&mockRoomService{
		deleteRoomFn: func(_ context.Context, _, _ string) error {
			return service.ErrRoomHasBookings
		},
	})

	router := setupRoomRouter(handler)
	req := httptest.NewRequest(http.MethodDelete, "/conferences/socrates-26/rooms/101", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "conflict")
}
