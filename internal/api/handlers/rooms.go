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

type serviceRoomService interface {
	ListRooms(ctx context.Context, slug string) ([]*service.RoomResponse, error)
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
