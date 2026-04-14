package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/service"
)

type serviceAttendeeService interface {
	ListAttendees(ctx context.Context, slug string, requesterUserID int64) (*service.AttendeeListResponse, error)
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
