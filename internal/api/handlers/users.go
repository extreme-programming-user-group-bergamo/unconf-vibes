package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/middleware"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/service"
)

type serviceUserService interface {
	GetProfile(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID int64, input service.UpdateProfileInput) (*models.User, error)
}

// UpdateProfileRequest is the JSON body for PUT /users/me.
type UpdateProfileRequest struct {
	DisplayName    *string `json:"display_name"`
	PrivacySetting *string `json:"privacy_setting"`
}

// UserHandler handles user profile endpoints.
type UserHandler struct {
	userService serviceUserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService serviceUserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetMe handles GET /users/me.
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			responses.WriteError(c, "user_not_found", "User not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to get user profile", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateMe handles PUT /users/me.
func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		responses.WriteError(c, "invalid_request", "Invalid request body", http.StatusBadRequest)
		return
	}

	input := service.UpdateProfileInput{
		DisplayName:    req.DisplayName,
		PrivacySetting: req.PrivacySetting,
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPrivacySetting) {
			responses.WriteError(c, "invalid_privacy_setting", err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, service.ErrUserNotFound) {
			responses.WriteError(c, "user_not_found", "User not found", http.StatusNotFound)
			return
		}

		slog.Error("failed to update user profile", "error", err, "user_id", userID)
		responses.WriteError(c, "internal_error", "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, user)
}
