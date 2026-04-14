package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/service"
)

type serviceConferenceService interface {
	ListConferences(ctx context.Context) ([]*service.ConferenceResponse, error)
	GetConference(ctx context.Context, slug string) (*service.ConferenceResponse, error)
	CreateConference(ctx context.Context, creatorUserID int64, input service.CreateConferenceInput) (*service.ConferenceResponse, error)
	UpdateConference(ctx context.Context, slug string, input service.UpdateConferenceInput) (*service.ConferenceResponse, error)
}

// ConferenceHandler handles conference API endpoints.
type ConferenceHandler struct {
	conferenceService serviceConferenceService
}

type CreateConferenceRequest struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Capacity    int    `json:"capacity"`
	HotelEmail  string `json:"hotel_email,omitempty"`
}

type UpdateConferenceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Capacity    int    `json:"capacity"`
	HotelEmail  string `json:"hotel_email,omitempty"`
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

// Create handles POST /conferences.
func (h *ConferenceHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req CreateConferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		responses.WriteError(c, "invalid_request", "Invalid start_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		responses.WriteError(c, "invalid_request", "Invalid end_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	created, err := h.conferenceService.CreateConference(c.Request.Context(), userID, service.CreateConferenceInput{
		Slug:        req.Slug,
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   startDate,
		EndDate:     endDate,
		Capacity:    req.Capacity,
		HotelEmail:  req.HotelEmail,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrganizerForbidden):
			responses.WriteError(c, "forbidden", "Organizer permissions required", http.StatusForbidden)
		case errors.Is(err, service.ErrInvalidConferenceInput):
			responses.WriteError(c, "invalid_request", "Invalid conference payload", http.StatusBadRequest)
		case errors.Is(err, service.ErrConferenceExists):
			responses.WriteError(c, "conflict", "Conference slug already exists", http.StatusConflict)
		default:
			slog.Error("failed to create conference", "error", err, "slug", req.Slug, "user_id", userID)
			responses.WriteError(c, "internal_error", "Failed to create conference", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Update handles PUT /conferences/:slug.
func (h *ConferenceHandler) Update(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		responses.WriteError(c, "invalid_request", "Conference slug is required", http.StatusBadRequest)
		return
	}

	var req UpdateConferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.StartDate))
	if err != nil {
		responses.WriteError(c, "invalid_request", "Invalid start_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}
	endDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EndDate))
	if err != nil {
		responses.WriteError(c, "invalid_request", "Invalid end_date format (expected YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	updated, err := h.conferenceService.UpdateConference(c.Request.Context(), slug, service.UpdateConferenceInput{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		StartDate:   startDate,
		EndDate:     endDate,
		Capacity:    req.Capacity,
		HotelEmail:  req.HotelEmail,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConferenceNotFound):
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidConferenceInput):
			responses.WriteError(c, "invalid_request", "Invalid conference payload", http.StatusBadRequest)
		default:
			slog.Error("failed to update conference", "error", err, "slug", slug)
			responses.WriteError(c, "internal_error", "Failed to update conference", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}
