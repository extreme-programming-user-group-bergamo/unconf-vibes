package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/service"
)

type serviceAttendeeService interface {
	ListAttendees(ctx context.Context, slug string, requesterUserID int64) (*service.AttendeeListResponse, error)
	GetOrganizerDashboard(ctx context.Context, slug string, requesterUserID int64, filters service.OrganizerDashboardFilters) (*service.OrganizerDashboardResponse, error)
}

// AttendeeHandler handles attendee API endpoints.
type AttendeeHandler struct {
	attendeeService serviceAttendeeService
}

// NewAttendeeHandler creates a new AttendeeHandler.
func NewAttendeeHandler(attendeeService serviceAttendeeService) *AttendeeHandler {
	return &AttendeeHandler{attendeeService: attendeeService}
}

// ListByConference handles GET /conferences/:slug/attendees.
func (h *AttendeeHandler) ListByConference(c *gin.Context) {
	slug := c.Param("slug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	result, err := h.attendeeService.ListAttendees(c.Request.Context(), slug, userID)
	if err != nil {
		if errors.Is(err, service.ErrConferenceNotFound) {
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to list attendees", "error", err, "slug", slug)
		responses.WriteError(c, "internal_error", "Failed to list attendees", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, result)
}

// OrganizerDashboard handles GET /conferences/:slug/dashboard.
func (h *AttendeeHandler) OrganizerDashboard(c *gin.Context) {
	slug := c.Param("slug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	hasSpecialRequests, err := parseBoolQueryParam(c.Query("has_special_requests"))
	if err != nil {
		responses.WriteError(c, "invalid_request", "has_special_requests must be true or false", http.StatusBadRequest)
		return
	}

	result, err := h.attendeeService.GetOrganizerDashboard(
		c.Request.Context(),
		slug,
		userID,
		service.OrganizerDashboardFilters{
			RoomType:           strings.TrimSpace(c.Query("room_type")),
			BookingStatus:      strings.TrimSpace(c.Query("booking_status")),
			HasSpecialRequests: hasSpecialRequests,
			Search:             strings.TrimSpace(c.Query("q")),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConferenceNotFound):
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
		case errors.Is(err, service.ErrOrganizerForbidden):
			responses.WriteError(c, "forbidden", "Organizer permissions required", http.StatusForbidden)
		default:
			slog.Error("failed to load organizer dashboard", "error", err, "slug", slug)
			responses.WriteError(c, "internal_error", "Failed to load organizer dashboard", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseBoolQueryParam(raw string) (bool, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}

	return parsed, nil
}
