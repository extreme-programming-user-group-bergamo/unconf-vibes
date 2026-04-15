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
	"github.com/katurdays/unconf/internal/service"
)

type serviceBookingService interface {
	CreateBooking(ctx context.Context, userID int64, input service.CreateBookingInput) (*service.BookingResponse, error)
	ListBookings(ctx context.Context, userID int64) ([]service.BookingResponse, error)
	CancelBooking(ctx context.Context, userID int64, bookingID int64) (*service.BookingResponse, error)
}

type createBookingRequest struct {
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	PrivacySetting string `json:"privacy_setting"`
	Notes          string `json:"notes,omitempty"`
}

type BookingHandler struct {
	bookingService serviceBookingService
}

func NewBookingHandler(bookingService serviceBookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

func (h *BookingHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req createBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.RoomID <= 0 || req.ConferenceID <= 0 {
		responses.WriteError(c, "invalid_request", "room_id and conference_id must be positive", http.StatusBadRequest)
		return
	}

	created, err := h.bookingService.CreateBooking(c.Request.Context(), userID, service.CreateBookingInput{
		RoomID:         req.RoomID,
		ConferenceID:   req.ConferenceID,
		PrivacySetting: req.PrivacySetting,
		Notes:          req.Notes,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound), errors.Is(err, service.ErrConferenceNotFound):
			responses.WriteError(c, "not_found", "Room not found", http.StatusNotFound)
		case errors.Is(err, service.ErrRoomFull):
			responses.WriteError(c, "room_full", "Room is full", http.StatusConflict)
		case errors.Is(err, service.ErrAlreadyBooked):
			responses.WriteError(c, "already_booked", "Booking already exists for this conference", http.StatusConflict)
		case errors.Is(err, service.ErrInvalidPrivacySetting):
			responses.WriteError(c, "invalid_request", "Invalid privacy setting", http.StatusBadRequest)
		default:
			slog.Error("failed to create booking", "error", err, "user_id", userID, "room_id", req.RoomID, "conference_id", req.ConferenceID)
			responses.WriteError(c, "internal_error", "Failed to create booking", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *BookingHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	bookings, err := h.bookingService.ListBookings(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list bookings", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to list bookings", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || bookingID <= 0 {
		responses.WriteError(c, "invalid_request", "Invalid booking ID", http.StatusBadRequest)
		return
	}

	cancelled, err := h.bookingService.CancelBooking(c.Request.Context(), userID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			responses.WriteError(c, "not_found", "Booking not found", http.StatusNotFound)
		case errors.Is(err, service.ErrBookingForbidden):
			responses.WriteError(c, "forbidden", "You cannot cancel this booking", http.StatusForbidden)
		default:
			slog.Error("failed to cancel booking", "error", err, "user_id", userID, "booking_id", bookingID)
			responses.WriteError(c, "internal_error", "Failed to cancel booking", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, cancelled)
}
