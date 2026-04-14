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

type mockAttendeeService struct {
	listAttendeesFn func(ctx context.Context, slug string, requesterUserID int64) (*service.AttendeeListResponse, error)
	dashboardFn     func(ctx context.Context, slug string, requesterUserID int64, filters service.OrganizerDashboardFilters) (*service.OrganizerDashboardResponse, error)
}

func (m *mockAttendeeService) ListAttendees(ctx context.Context, slug string, requesterUserID int64) (*service.AttendeeListResponse, error) {
	if m.listAttendeesFn != nil {
		return m.listAttendeesFn(ctx, slug, requesterUserID)
	}

	return &service.AttendeeListResponse{}, nil
}

func (m *mockAttendeeService) GetOrganizerDashboard(
	ctx context.Context,
	slug string,
	requesterUserID int64,
	filters service.OrganizerDashboardFilters,
) (*service.OrganizerDashboardResponse, error) {
	if m.dashboardFn != nil {
		return m.dashboardFn(ctx, slug, requesterUserID, filters)
	}
	return &service.OrganizerDashboardResponse{}, nil
}

func setupAttendeeRouter(handler *AttendeeHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", int64(42))
		c.Next()
	})
	router.GET("/conferences/:slug/attendees", handler.ListByConference)
	router.GET("/conferences/:slug/dashboard", handler.OrganizerDashboard)
	return router
}

func TestAttendeeHandler_ListByConference_Success(t *testing.T) {
	handler := NewAttendeeHandler(&mockAttendeeService{
		listAttendeesFn: func(_ context.Context, slug string, requesterUserID int64) (*service.AttendeeListResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, int64(42), requesterUserID)
			return &service.AttendeeListResponse{
				Attendees: []service.AttendeeResponse{
					{
						DisplayName:    "Alice",
						GitHubUsername: "alice",
						Room: &service.AttendeeRoomResponse{
							RoomNumber: "101",
							RoomType:   "double",
						},
					},
				},
				PrivateAttendeesCount: 2,
			}, nil
		},
	})

	router := setupAttendeeRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/attendees", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"display_name":"Alice"`)
	assert.Contains(t, w.Body.String(), `"private_attendees_count":2`)
}

func TestAttendeeHandler_ListByConference_ConferenceNotFound(t *testing.T) {
	handler := NewAttendeeHandler(&mockAttendeeService{
		listAttendeesFn: func(_ context.Context, _ string, _ int64) (*service.AttendeeListResponse, error) {
			return nil, service.ErrConferenceNotFound
		},
	})

	router := setupAttendeeRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/nonexistent/attendees", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
}

func TestAttendeeHandler_ListByConference_ServiceError(t *testing.T) {
	handler := NewAttendeeHandler(&mockAttendeeService{
		listAttendeesFn: func(_ context.Context, _ string, _ int64) (*service.AttendeeListResponse, error) {
			return nil, errors.New("db error")
		},
	})

	router := setupAttendeeRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/attendees", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

func TestAttendeeHandler_OrganizerDashboard_Success(t *testing.T) {
	handler := NewAttendeeHandler(&mockAttendeeService{
		dashboardFn: func(_ context.Context, slug string, requesterUserID int64, filters service.OrganizerDashboardFilters) (*service.OrganizerDashboardResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, int64(42), requesterUserID)
			assert.Equal(t, "double", filters.RoomType)
			assert.Equal(t, "confirmed", filters.BookingStatus)
			assert.True(t, filters.HasSpecialRequests)
			assert.Equal(t, "ali", filters.Search)
			return &service.OrganizerDashboardResponse{
				ConferenceSlug:     "socrates-26",
				TotalRegistrations: 2,
				CapacityUsagePct:   50,
				Attendees: []service.OrganizerDashboardAttendeeResponse{
					{Name: "Alice", Email: "alice@test.dev", PrivacySetting: "private", BookingStatus: "confirmed"},
				},
			}, nil
		},
	})

	router := setupAttendeeRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/dashboard?room_type=double&booking_status=confirmed&has_special_requests=true&q=ali", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"conference_slug":"socrates-26"`)
	assert.Contains(t, w.Body.String(), `"capacity_usage_pct":50`)
	assert.Contains(t, w.Body.String(), `"email":"alice@test.dev"`)
}

func TestAttendeeHandler_OrganizerDashboard_InvalidBool(t *testing.T) {
	handler := NewAttendeeHandler(&mockAttendeeService{})
	router := setupAttendeeRouter(handler)
	req := httptest.NewRequest(http.MethodGet, "/conferences/socrates-26/dashboard?has_special_requests=not-bool", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}
