package service

import (
	"errors"
	"fmt"
)

var (
	ErrAuthorizationPending  = errors.New("authorization pending")
	ErrSlowDown              = errors.New("authorization slow down")
	ErrAccessDenied          = errors.New("authorization denied")
	ErrExpiredDeviceCode     = errors.New("device code expired")
	ErrInvalidDeviceCode     = errors.New("invalid device code")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrExpiredRefreshToken   = errors.New("refresh token expired")
	ErrRevokedRefreshToken   = errors.New("refresh token revoked")
	ErrUserNotFound          = errors.New("user not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrInvalidPrivacySetting = errors.New("invalid privacy setting: must be public, private, or connections_only")
	ErrConferenceNotFound    = errors.New("conference not found")
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
