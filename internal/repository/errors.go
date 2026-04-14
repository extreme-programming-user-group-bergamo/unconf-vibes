package repository

import "errors"

var (
	ErrUserNotFound                = errors.New("user not found")
	ErrUserExists                  = errors.New("user already exists")
	ErrDatabaseInit                = errors.New("failed to initialize database")
	ErrRefreshSessionNotFound      = errors.New("refresh session not found")
	ErrConferenceNotFound          = errors.New("conference not found")
	ErrConferenceExists            = errors.New("conference already exists")
	ErrConferenceOrganizerNotFound = errors.New("conference organizer not found")
	ErrConferenceOrganizerExists   = errors.New("conference organizer already exists")
	ErrRoomNotFound                = errors.New("room not found")
	ErrRoomExists                  = errors.New("room already exists")
	ErrRoomHasBookings             = errors.New("room has bookings")
	ErrBookingNotFound             = errors.New("booking not found")
	ErrBookingExists               = errors.New("booking already exists")
	ErrEmailLogNotFound            = errors.New("email log not found")
	ErrRoommateRequestNotFound     = errors.New("roommate request not found")
	ErrRoommateRequestExists       = errors.New("roommate request already exists")
)
