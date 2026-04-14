package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
)

type serviceRoomService interface {
	ListRooms(ctx context.Context, slug string) ([]*service.RoomResponse, error)
	CreateRoom(ctx context.Context, slug string, input service.ManageRoomInput) (*models.Room, error)
	UpdateRoom(ctx context.Context, slug string, roomNumber string, input service.ManageRoomInput) (*models.Room, error)
	DeleteRoom(ctx context.Context, slug string, roomNumber string) error
}

type upsertRoomRequest struct {
	RoomNumber    string  `json:"room_number"`
	RoomType      string  `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
	Capacity      int     `json:"capacity"`
}

// RoomHandler handles room API endpoints.
type RoomHandler struct {
	roomService serviceRoomService
}

// NewRoomHandler creates a new RoomHandler.
func NewRoomHandler(roomService serviceRoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// ListByConference handles GET /conferences/:slug/rooms.
func (h *RoomHandler) ListByConference(c *gin.Context) {
	slug := c.Param("slug")

	rooms, err := h.roomService.ListRooms(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, service.ErrConferenceNotFound) {
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to list rooms", "error", err, "slug", slug)
		responses.WriteError(c, "internal_error", "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, rooms)
}

// Create handles POST /conferences/:slug/rooms.
func (h *RoomHandler) Create(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		responses.WriteError(c, "invalid_request", "Conference slug is required", http.StatusBadRequest)
		return
	}

	var req upsertRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.roomService.CreateRoom(c.Request.Context(), slug, service.ManageRoomInput{
		RoomNumber:    req.RoomNumber,
		RoomType:      req.RoomType,
		PricePerNight: req.PricePerNight,
		Capacity:      req.Capacity,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConferenceNotFound):
			responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidRoomInput):
			responses.WriteError(c, "invalid_request", "Invalid room payload", http.StatusBadRequest)
		case errors.Is(err, service.ErrRoomExists):
			responses.WriteError(c, "conflict", "Room number already exists", http.StatusConflict)
		default:
			slog.Error("failed to create room", "error", err, "slug", slug)
			responses.WriteError(c, "internal_error", "Failed to create room", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, created)
}

// Update handles PUT /conferences/:slug/rooms/:number.
func (h *RoomHandler) Update(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	roomNumber := strings.TrimSpace(c.Param("number"))
	if slug == "" || roomNumber == "" {
		responses.WriteError(c, "invalid_request", "Conference slug and room number are required", http.StatusBadRequest)
		return
	}

	var req upsertRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.roomService.UpdateRoom(c.Request.Context(), slug, roomNumber, service.ManageRoomInput{
		RoomNumber:    req.RoomNumber,
		RoomType:      req.RoomType,
		PricePerNight: req.PricePerNight,
		Capacity:      req.Capacity,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConferenceNotFound), errors.Is(err, service.ErrRoomNotFound):
			responses.WriteError(c, "not_found", "Room not found", http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidRoomInput):
			responses.WriteError(c, "invalid_request", "Invalid room payload", http.StatusBadRequest)
		case errors.Is(err, service.ErrRoomExists):
			responses.WriteError(c, "conflict", "Room number already exists", http.StatusConflict)
		default:
			slog.Error("failed to update room", "error", err, "slug", slug, "room_number", roomNumber)
			responses.WriteError(c, "internal_error", "Failed to update room", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

// Delete handles DELETE /conferences/:slug/rooms/:number.
func (h *RoomHandler) Delete(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	roomNumber := strings.TrimSpace(c.Param("number"))
	if slug == "" || roomNumber == "" {
		responses.WriteError(c, "invalid_request", "Conference slug and room number are required", http.StatusBadRequest)
		return
	}

	if err := h.roomService.DeleteRoom(c.Request.Context(), slug, roomNumber); err != nil {
		switch {
		case errors.Is(err, service.ErrConferenceNotFound), errors.Is(err, service.ErrRoomNotFound):
			responses.WriteError(c, "not_found", "Room not found", http.StatusNotFound)
		case errors.Is(err, service.ErrInvalidRoomInput):
			responses.WriteError(c, "invalid_request", "Invalid room number", http.StatusBadRequest)
		case errors.Is(err, service.ErrRoomHasBookings):
			responses.WriteError(c, "conflict", "Room has bookings and cannot be removed", http.StatusConflict)
		default:
			slog.Error("failed to delete room", "error", err, "slug", slug, "room_number", roomNumber)
			responses.WriteError(c, "internal_error", "Failed to delete room", http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
