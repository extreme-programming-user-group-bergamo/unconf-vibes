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

type serviceRequestService interface {
	CreateRequest(ctx context.Context, requesterID int64, input service.CreateRoommateRequestInput) (*models.RoommateRequest, error)
	ListRequests(ctx context.Context, userID int64) ([]service.RoommateRequestView, error)
	AcceptRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error)
	DeclineRequest(ctx context.Context, requestID int64, userID int64) (*models.RoommateRequest, error)
}

type CreateRoommateRequestRequest struct {
	TargetID       int64  `json:"target_id,omitempty"`
	TargetUsername string `json:"target_username,omitempty"`
	RoomID         int64  `json:"room_id"`
}

type RequestHandler struct {
	requestService serviceRequestService
}

func NewRequestHandler(requestService serviceRequestService) *RequestHandler {
	return &RequestHandler{requestService: requestService}
}

func (h *RequestHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req CreateRoommateRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.requestService.CreateRequest(c.Request.Context(), userID, service.CreateRoommateRequestInput{
		TargetID:       req.TargetID,
		TargetUsername: req.TargetUsername,
		RoomID:         req.RoomID,
	})
	if err != nil {
		h.writeServiceError(c, err, "failed to create roommate request")
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *RequestHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	result, err := h.requestService.ListRequests(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list roommate requests", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to list roommate requests", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *RequestHandler) Accept(c *gin.Context) {
	h.respondToRequest(c, true)
}

func (h *RequestHandler) Decline(c *gin.Context) {
	h.respondToRequest(c, false)
}

func (h *RequestHandler) respondToRequest(c *gin.Context, accept bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || requestID <= 0 {
		responses.WriteError(c, "invalid_request", "Invalid request ID", http.StatusBadRequest)
		return
	}

	var updated *models.RoommateRequest
	if accept {
		updated, err = h.requestService.AcceptRequest(c.Request.Context(), requestID, userID)
	} else {
		updated, err = h.requestService.DeclineRequest(c.Request.Context(), requestID, userID)
	}
	if err != nil {
		h.writeServiceError(c, err, "failed to update roommate request")
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *RequestHandler) writeServiceError(c *gin.Context, err error, logMessage string) {
	switch {
	case errors.Is(err, service.ErrCannotRequestSelf):
		responses.WriteError(c, "cannot_request_self", "Cannot create roommate request for yourself", http.StatusBadRequest)
	case errors.Is(err, service.ErrRoomFull):
		responses.WriteError(c, "room_full", "Room is full", http.StatusConflict)
	case errors.Is(err, service.ErrRoomNotFound):
		responses.WriteError(c, "not_found", "Room not found", http.StatusNotFound)
	case errors.Is(err, service.ErrRequesterNotInRoom):
		responses.WriteError(c, "invalid_request", "Requester is not in this room", http.StatusBadRequest)
	case errors.Is(err, service.ErrDuplicateRequest):
		responses.WriteError(c, "duplicate_request", "Duplicate roommate request", http.StatusConflict)
	case errors.Is(err, service.ErrTargetNotFound):
		responses.WriteError(c, "target_not_found", "Target user not found", http.StatusNotFound)
	case errors.Is(err, service.ErrTargetAlreadyBooked):
		responses.WriteError(c, "target_has_booking", "Target user already has a booking for this conference", http.StatusConflict)
	case errors.Is(err, service.ErrRequestNotFound):
		responses.WriteError(c, "not_found", "Roommate request not found", http.StatusNotFound)
	case errors.Is(err, service.ErrRequestForbidden):
		responses.WriteError(c, "forbidden", "You cannot modify this roommate request", http.StatusForbidden)
	case errors.Is(err, service.ErrInvalidRequestState):
		responses.WriteError(c, "invalid_request_state", "Roommate request is already resolved", http.StatusConflict)
	default:
		slog.Error(logMessage, "error", err)
		responses.WriteError(c, "internal_error", "Internal server error", http.StatusInternalServerError)
	}
}
