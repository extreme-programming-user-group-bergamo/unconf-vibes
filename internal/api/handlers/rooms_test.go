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

type mockRoomService struct {
	listRoomsFn func(ctx context.Context, slug string) ([]*service.RoomResponse, error)
}

func (m *mockRoomService) ListRooms(ctx context.Context, slug string) ([]*service.RoomResponse, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx, slug)
	}
	return nil, nil
}

func setupRoomRouter(handler *RoomHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/conferences/:slug/rooms", handler.ListByConference)
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
