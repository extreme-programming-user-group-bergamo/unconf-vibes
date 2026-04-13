package models

import "time"

type Room struct {
	ID            int64     `json:"id" db:"id"`
	ConferenceID  int64     `json:"conference_id" db:"conference_id"`
	RoomNumber    string    `json:"room_number" db:"room_number"`
	RoomType      string    `json:"room_type" db:"room_type"`
	PricePerNight float64   `json:"price_per_night" db:"price_per_night"`
	Capacity      int       `json:"capacity" db:"capacity"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
