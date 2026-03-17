package repository

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserExists             = errors.New("user already exists")
	ErrDatabaseInit           = errors.New("failed to initialize database")
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
	ErrConferenceNotFound     = errors.New("conference not found")
	ErrConferenceExists       = errors.New("conference already exists")
)
