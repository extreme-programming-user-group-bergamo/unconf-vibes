package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
)

type serviceOrganizerService interface {
	AddOrganizer(ctx context.Context, slug string, requesterUserID int64, targetUserID int64, role string) (*models.ConferenceOrganizer, error)
	RemoveOrganizer(ctx context.Context, slug string, requesterUserID int64, targetUserID int64) error
}

type OrganizerHandler struct {
	organizerService serviceOrganizerService
}

type AddOrganizerRequest struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

func NewOrganizerHandler(organizerService serviceOrganizerService) *OrganizerHandler {
	return &OrganizerHandler{organizerService: organizerService}
}

func (h *OrganizerHandler) Add(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req AddOrganizerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.UserID <= 0 {
		responses.WriteError(c, "invalid_request", "user_id must be positive", http.StatusBadRequest)
		return
	}

	created, err := h.organizerService.AddOrganizer(c.Request.Context(), c.Param("slug"), userID, req.UserID, req.Role)
	if err != nil {
		h.writeServiceError(c, err, "failed to add organizer")
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *OrganizerHandler) Remove(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	targetUserID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil || targetUserID <= 0 {
		responses.WriteError(c, "invalid_request", "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.organizerService.RemoveOrganizer(c.Request.Context(), c.Param("slug"), userID, targetUserID); err != nil {
		h.writeServiceError(c, err, "failed to remove organizer")
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OrganizerHandler) writeServiceError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, service.ErrConferenceNotFound), errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrOrganizerNotFound):
		responses.WriteError(c, "not_found", "Resource not found", http.StatusNotFound)
	case errors.Is(err, service.ErrOrganizerForbidden), errors.Is(err, service.ErrOwnerRequired):
		responses.WriteError(c, "forbidden", "Owner permissions required", http.StatusForbidden)
	case errors.Is(err, service.ErrOrganizerExists):
		responses.WriteError(c, "conflict", "Organizer already exists", http.StatusConflict)
	default:
		slog.Error(logMessage, "error", err)
		responses.WriteError(c, "internal_error", "Internal server error", http.StatusInternalServerError)
	}
}
