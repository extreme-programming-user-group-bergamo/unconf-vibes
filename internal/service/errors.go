package service

import (
	"errors"
	"fmt"
)

var (
	ErrAuthorizationPending     = errors.New("authorization pending")
	ErrSlowDown                 = errors.New("authorization slow down")
	ErrAccessDenied             = errors.New("authorization denied")
	ErrExpiredDeviceCode        = errors.New("device code expired")
	ErrInvalidDeviceCode        = errors.New("invalid device code")
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrExpiredRefreshToken      = errors.New("refresh token expired")
	ErrRevokedRefreshToken      = errors.New("refresh token revoked")
	ErrUserNotFound             = errors.New("user not found")
	ErrSessionNotFound          = errors.New("session not found")
	ErrInvalidPrivacySetting    = errors.New("invalid privacy setting: must be public, private, or connections_only")
	ErrConferenceNotFound       = errors.New("conference not found")
	ErrConferenceExists         = errors.New("conference already exists")
	ErrInvalidConferenceInput   = errors.New("invalid conference input")
	ErrRoomNotFound             = errors.New("room not found")
	ErrRoomExists               = errors.New("room already exists")
	ErrRoomHasBookings          = errors.New("room has existing bookings")
	ErrInvalidRoomInput         = errors.New("invalid room input")
	ErrRoomFull                 = errors.New("room is full")
	ErrAlreadyBooked            = errors.New("user already has a booking for this conference")
	ErrBookingNotFound          = errors.New("booking not found")
	ErrBookingForbidden         = errors.New("booking is not owned by user")
	ErrCannotRequestSelf        = errors.New("cannot request yourself")
	ErrRequestNotFound          = errors.New("roommate request not found")
	ErrRequestForbidden         = errors.New("roommate request is not owned by user")
	ErrInvalidRequestState      = errors.New("roommate request is not pending")
	ErrRequesterNotInRoom       = errors.New("requester does not occupy room")
	ErrDuplicateRequest         = errors.New("duplicate roommate request")
	ErrTargetNotFound           = errors.New("target user not found")
	ErrTargetAlreadyBooked      = errors.New("target user already has booking")
	ErrOrganizerForbidden       = errors.New("organizer permissions required")
	ErrOrganizerNotFound        = errors.New("organizer not found")
	ErrOrganizerExists          = errors.New("organizer already exists")
	ErrOwnerRequired            = errors.New("owner role required")
	ErrHotelEmailNotConfigured  = errors.New("hotel email not configured")
	ErrHotelEmailInvalid        = errors.New("hotel email invalid")
	ErrHotelEmailDeliveryFailed = errors.New("hotel email delivery failed")
)

type PendingAuthError struct {
	Interval int
	Cause    error
}

func (e *PendingAuthError) Error() string {
	if e.Interval > 0 {
		return fmt.Sprintf("authorization pending; retry_interval=%d", e.Interval)
	}

	return "authorization pending"
}

func (e *PendingAuthError) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}

	return ErrAuthorizationPending
}
