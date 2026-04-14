package models

import "time"

type ConferenceOrganizerRole string

const (
	ConferenceOrganizerRoleOwner ConferenceOrganizerRole = "owner"
	ConferenceOrganizerRoleAdmin ConferenceOrganizerRole = "admin"
)

type ConferenceOrganizer struct {
	ID           int64                   `json:"id" db:"id"`
	ConferenceID int64                   `json:"conference_id" db:"conference_id"`
	UserID       int64                   `json:"user_id" db:"user_id"`
	Role         ConferenceOrganizerRole `json:"role" db:"role"`
	CreatedAt    time.Time               `json:"created_at" db:"created_at"`
}
