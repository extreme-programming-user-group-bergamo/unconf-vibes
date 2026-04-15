package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBookingService struct {
	createFn func(context.Context, int64, service.CreateBookingInput) (*service.BookingResponse, error)
	listFn   func(context.Context, int64) ([]service.BookingResponse, error)
	cancelFn func(context.Context, int64, int64) (*service.BookingResponse, error)
}

func (m *mockBookingService) CreateBooking(ctx context.Context, userID int64, input service.CreateBookingInput) (*service.BookingResponse, error) {
	if m.createFn == nil {
		return nil, nil
	}

	return m.createFn(ctx, userID, input)
}

func (m *mockBookingService) ListBookings(ctx context.Context, userID int64) ([]service.BookingResponse, error) {
	return m.listFn(ctx, userID)
}

func (m *mockBookingService) CancelBooking(ctx context.Context, userID int64, bookingID int64) (*service.BookingResponse, error) {
	return m.cancelFn(ctx, userID, bookingID)
}

func TestBookingHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
		createFn: func(_ context.Context, userID int64, input service.CreateBookingInput) (*service.BookingResponse, error) {
			assert.Equal(t, int64(9), userID)
			assert.Equal(t, int64(22), input.RoomID)
			assert.Equal(t, int64(3), input.ConferenceID)
			assert.Equal(t, "private", input.PrivacySetting)
			assert.Equal(t, "late arrival", input.Notes)
			return &service.BookingResponse{ID: 44, Status: "requested", PrivacySetting: "private", Notes: "late arrival"}, nil
		},
		listFn: func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
		cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) {
			return nil, nil
		},
	})

	router := gin.New()
	router.POST("/bookings", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{"room_id":22,"conference_id":3,"privacy_setting":"private","notes":"late arrival"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"requested"`)
}

func TestBookingHandler_Create_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
		listFn: func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
		cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) {
			return nil, nil
		},
	})

	router := gin.New()
	router.POST("/bookings", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{"privacy_setting":"private"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBookingHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
		createFn: func(_ context.Context, _ int64, _ service.CreateBookingInput) (*service.BookingResponse, error) {
			return nil, nil
		},
		listFn: func(_ context.Context, userID int64) ([]service.BookingResponse, error) {
			assert.Equal(t, int64(9), userID)
			return []service.BookingResponse{{ID: 1, Status: "confirmed"}}, nil
		},
		cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) {
			return nil, nil
		},
	})

	router := gin.New()
	router.GET("/bookings", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/bookings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"confirmed"`)
}

func TestBookingHandler_Cancel_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
		createFn: func(_ context.Context, _ int64, _ service.CreateBookingInput) (*service.BookingResponse, error) {
			return nil, nil
		},
		listFn: func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
		cancelFn: func(_ context.Context, userID, bookingID int64) (*service.BookingResponse, error) {
			assert.Equal(t, int64(9), userID)
			assert.Equal(t, int64(22), bookingID)
			return &service.BookingResponse{ID: 22, Status: "cancelled"}, nil
		},
	})

	router := gin.New()
	router.DELETE("/bookings/:id", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Cancel(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/bookings/22", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"cancelled"`)
}

func TestBookingHandler_Cancel_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
		createFn: func(_ context.Context, _ int64, _ service.CreateBookingInput) (*service.BookingResponse, error) {
			return nil, nil
		},
		listFn:   func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
		cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) { return nil, nil },
	})

	router := gin.New()
	router.DELETE("/bookings/:id", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Cancel(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/bookings/nope", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBookingHandler_Create_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
		code       string
	}{
		{name: "not found", err: service.ErrRoomNotFound, statusCode: http.StatusNotFound, code: "not_found"},
		{name: "room full", err: service.ErrRoomFull, statusCode: http.StatusConflict, code: "room_full"},
		{name: "already booked", err: service.ErrAlreadyBooked, statusCode: http.StatusConflict, code: "already_booked"},
		{name: "invalid privacy", err: service.ErrInvalidPrivacySetting, statusCode: http.StatusBadRequest, code: "invalid_request"},
		{name: "internal", err: errors.New("boom"), statusCode: http.StatusInternalServerError, code: "internal_error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			handler := NewBookingHandler(&mockBookingService{
				createFn: func(_ context.Context, _ int64, _ service.CreateBookingInput) (*service.BookingResponse, error) {
					return nil, tc.err
				},
				listFn: func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
				cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) {
					return nil, nil
				},
			})

			router := gin.New()
			router.POST("/bookings", func(c *gin.Context) {
				c.Set("user_id", int64(9))
				handler.Create(c)
			})

			req := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{"room_id":22,"conference_id":3,"privacy_setting":"private"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.statusCode, w.Code)
			assert.Contains(t, w.Body.String(), `"code":"`+tc.code+`"`)
		})
	}
}

func TestBookingHandler_Cancel_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{name: "not found", err: service.ErrBookingNotFound, statusCode: http.StatusNotFound},
		{name: "forbidden", err: service.ErrBookingForbidden, statusCode: http.StatusForbidden},
		{name: "internal", err: errors.New("boom"), statusCode: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			handler := NewBookingHandler(&mockBookingService{
				createFn: func(_ context.Context, _ int64, _ service.CreateBookingInput) (*service.BookingResponse, error) {
					return nil, nil
				},
				listFn: func(_ context.Context, _ int64) ([]service.BookingResponse, error) { return nil, nil },
				cancelFn: func(_ context.Context, _, _ int64) (*service.BookingResponse, error) {
					return nil, tc.err
				},
			})

			router := gin.New()
			router.DELETE("/bookings/:id", func(c *gin.Context) {
				c.Set("user_id", int64(9))
				handler.Cancel(c)
			})

			req := httptest.NewRequest(http.MethodDelete, "/bookings/22", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tc.statusCode, w.Code)
		})
	}
}
