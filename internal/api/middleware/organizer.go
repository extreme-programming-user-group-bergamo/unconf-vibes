package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/katurdays/unconf/internal/api/responses"
	"github.com/katurdays/unconf/internal/service"
)

type OrganizerPermissionChecker interface {
	IsOrganizerForConference(ctx context.Context, slug string, userID int64) (bool, error)
	IsOwnerForConference(ctx context.Context, slug string, userID int64) (bool, error)
	IsOrganizer(ctx context.Context, userID int64) (bool, error)
}

func RequireAnyOrganizer(checker OrganizerPermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
			c.Abort()
			return
		}

		isOrganizer, err := checker.IsOrganizer(c.Request.Context(), userID)
		if err != nil {
			responses.WriteError(c, "internal_error", "Failed to check organizer permissions", http.StatusInternalServerError)
			c.Abort()
			return
		}
		if !isOrganizer {
			responses.WriteError(c, "forbidden", "Organizer permissions required", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireOrganizer(checker OrganizerPermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
			c.Abort()
			return
		}

		isOrganizer, err := checker.IsOrganizerForConference(c.Request.Context(), c.Param("slug"), userID)
		if err != nil {
			if errors.Is(err, service.ErrConferenceNotFound) {
				responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
				c.Abort()
				return
			}
			responses.WriteError(c, "internal_error", "Failed to check organizer permissions", http.StatusInternalServerError)
			c.Abort()
			return
		}
		if !isOrganizer {
			responses.WriteError(c, "forbidden", "Organizer permissions required", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireOrganizerOwner(checker OrganizerPermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			responses.WriteError(c, "unauthorized", "User ID not found in context", http.StatusUnauthorized)
			c.Abort()
			return
		}

		isOwner, err := checker.IsOwnerForConference(c.Request.Context(), c.Param("slug"), userID)
		if err != nil {
			if errors.Is(err, service.ErrConferenceNotFound) {
				responses.WriteError(c, "not_found", "Conference not found", http.StatusNotFound)
				c.Abort()
				return
			}
			responses.WriteError(c, "internal_error", "Failed to check organizer permissions", http.StatusInternalServerError)
			c.Abort()
			return
		}
		if !isOwner {
			responses.WriteError(c, "forbidden", "Owner permissions required", http.StatusForbidden)
			c.Abort()
			return
		}

		c.Next()
	}
}
