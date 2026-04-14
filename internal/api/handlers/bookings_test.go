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
	"github.com/stretchr/testify/require"
)

type mockBookingService struct {
	listFn   func(context.Context, int64) ([]service.BookingResponse, error)
	cancelFn func(context.Context, int64, int64) (*service.BookingResponse, error)
}

func (m *mockBookingService) ListBookings(ctx context.Context, userID int64) ([]service.BookingResponse, error) {
	return m.listFn(ctx, userID)
}

func (m *mockBookingService) CancelBooking(ctx context.Context, userID int64, bookingID int64) (*service.BookingResponse, error) {
	return m.cancelFn(ctx, userID, bookingID)
}

func TestBookingHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewBookingHandler(&mockBookingService{
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
