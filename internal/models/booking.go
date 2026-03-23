package models

import "time"

type BookingStatus string

const (
	BookingStatusRequested  BookingStatus = "requested"
	BookingStatusConfirmed  BookingStatus = "confirmed"
	BookingStatusCancelled  BookingStatus = "cancelled"
)

type Booking struct {
	ID             int64         `json:"id" db:"id"`
	RoomID         int64         `json:"room_id" db:"room_id"`
	UserID         int64         `json:"user_id" db:"user_id"`
	ConferenceID   int64         `json:"conference_id" db:"conference_id"`
	Status         BookingStatus `json:"status" db:"status"`
	PrivacySetting string        `json:"privacy_setting" db:"privacy_setting"`
	Notes          string        `json:"notes,omitempty" db:"notes"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
	ConfirmedAt    *time.Time    `json:"confirmed_at,omitempty" db:"confirmed_at"`
	CancelledAt    *time.Time    `json:"cancelled_at,omitempty" db:"cancelled_at"`
}
