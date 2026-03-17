package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/service"
)

type serviceConferenceService interface {
	ListConferences(ctx context.Context) ([]*service.ConferenceResponse, error)
	GetConference(ctx context.Context, slug string) (*service.ConferenceResponse, error)
}

// ConferenceHandler handles conference API endpoints.
type ConferenceHandler struct {
	conferenceService serviceConferenceService
}

// NewConferenceHandler creates a new ConferenceHandler.
func NewConferenceHandler(conferenceService serviceConferenceService) *ConferenceHandler {
	return &ConferenceHandler{conferenceService: conferenceService}
}

// List handles GET /conferences.
func (h *ConferenceHandler) List(c *gin.Context) {
	conferences, err := h.conferenceService.ListConferences(c.Request.Context())
	if err != nil {
		slog.Error("failed to list conferences", "error", err)
		responses.WriteError(c, "internal_error", "Failed to list conferences", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, conferences)
}

// GetBySlug handles GET /conferences/:slug.
func (h *ConferenceHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	conf, err := h.conferenceService.GetConference(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrConferenceNotFound) {
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get conference", "error", err, "slug", slug)
		responses.WriteError(c, "internal_error", "Failed to get conference", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, conf)
}
