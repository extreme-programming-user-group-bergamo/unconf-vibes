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
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRequestService struct {
	createFn  func(ctx context.Context, requesterID int64, input service.CreateRoommateRequestInput) (*models.RoommateRequest, error)
	listFn    func(ctx context.Context, userID int64) ([]service.RoommateRequestView, error)
	acceptFn  func(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error)
	declineFn func(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error)
}

func (m *mockRequestService) CreateRequest(ctx context.Context, requesterID int64, input service.CreateRoommateRequestInput) (*models.RoommateRequest, error) {
	return m.createFn(ctx, requesterID, input)
}
func (m *mockRequestService) ListRequests(ctx context.Context, userID int64) ([]service.RoommateRequestView, error) {
	return m.listFn(ctx, userID)
}
func (m *mockRequestService) AcceptRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error) {
	return m.acceptFn(ctx, requestID, userID)
}
func (m *mockRequestService) DeclineRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error) {
	return m.declineFn(ctx, requestID, userID)
}

func TestRequestHandler_Create_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		createFn: func(_ context.Context, requesterID int64, input service.CreateRoommateRequestInput) (*models.RoommateRequest, error) {
			assert.Equal(t, int64(9), requesterID)
			assert.Equal(t, int64(2), input.TargetID)
			assert.Equal(t, int64(3), input.RoomID)
			return &models.RoommateRequest{
				ID:          1,
				RequesterID: requesterID,
				TargetID:    input.TargetID,
				RoomID:      input.RoomID,
				Status:      models.RoommateRequestStatusPending,
				CreatedAt:   time.Now().UTC(),
			}, nil
		},
	})

	router := gin.New()
	router.POST("/requests", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/requests", bytes.NewBufferString(`{"target_id":2,"room_id":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"pending"`)
}

func TestRequestHandler_Create_CannotRequestSelf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		createFn: func(_ context.Context, _ int64, _ service.CreateRoommateRequestInput) (*models.RoommateRequest, error) {
			return nil, service.ErrCannotRequestSelf
		},
	})

	router := gin.New()
	router.POST("/requests", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/requests", bytes.NewBufferString(`{"target_id":9,"room_id":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "cannot_request_self")
}

func TestRequestHandler_Create_TargetNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		createFn: func(_ context.Context, _ int64, input service.CreateRoommateRequestInput) (*models.RoommateRequest, error) {
			assert.Equal(t, "missing", input.TargetUsername)
			return nil, service.ErrTargetNotFound
		},
	})

	router := gin.New()
	router.POST("/requests", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/requests", bytes.NewBufferString(`{"target_username":"missing","room_id":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "target_not_found")
}

func TestRequestHandler_Create_TargetHasBooking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		createFn: func(_ context.Context, _ int64, _ service.CreateRoommateRequestInput) (*models.RoommateRequest, error) {
			return nil, service.ErrTargetAlreadyBooked
		},
	})

	router := gin.New()
	router.POST("/requests", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.Create(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/requests", bytes.NewBufferString(`{"target_username":"booked","room_id":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "target_has_booking")
}

func TestRequestHandler_List_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		listFn: func(_ context.Context, userID int64) ([]service.RoommateRequestView, error) {
			assert.Equal(t, int64(9), userID)
			return []service.RoommateRequestView{
				{
					ID:            1,
					RequesterID:   9,
					TargetID:      2,
					RoomID:        3,
					RoomNumber:    "101",
					RoomType:      "double",
					ConferenceID:  5,
					Status:        models.RoommateRequestStatusPending,
					Direction:     "outgoing",
					RequesterName: "Requester",
					TargetName:    "Target",
					CreatedAt:     "2026-01-01T12:00:00Z",
				},
			}, nil
		},
	})

	router := gin.New()
	router.GET("/requests", func(c *gin.Context) {
		c.Set("user_id", int64(9))
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/requests", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"direction":"outgoing"`)
	assert.Contains(t, w.Body.String(), `"requester_name":"Requester"`)
	assert.Contains(t, w.Body.String(), `"target_name":"Target"`)
	assert.Contains(t, w.Body.String(), `"room_number":"101"`)
}

func TestRequestHandler_Accept_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		acceptFn: func(_ context.Context, _ int64, _ int64) (*models.RoommateRequest, error) {
			return nil, service.ErrRequestForbidden
		},
	})

	router := gin.New()
	router.PUT("/requests/:id/accept", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		handler.Accept(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/requests/5/accept", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestRequestHandler_Decline_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{
		declineFn: func(_ context.Context, _ int64, _ int64) (*models.RoommateRequest, error) {
			return nil, errors.New("boom")
		},
	})

	router := gin.New()
	router.PUT("/requests/:id/decline", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		handler.Decline(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/requests/5/decline", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

func TestRequestHandler_Accept_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRequestHandler(&mockRequestService{})

	router := gin.New()
	router.PUT("/requests/:id/accept", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		handler.Accept(c)
	})

	req := httptest.NewRequest(http.MethodPut, "/requests/not-a-number/accept", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}
