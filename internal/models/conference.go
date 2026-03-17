package models

import "time"

type ConferenceStatus string

const (
	ConferenceStatusUpcoming ConferenceStatus = "upcoming"
	ConferenceStatusActive   ConferenceStatus = "active"
	ConferenceStatusPast     ConferenceStatus = "past"
)

type Conference struct {
	ID          int64     `json:"id" db:"id"`
	Slug        string    `json:"slug" db:"slug"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Location    string    `json:"location" db:"location"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	EndDate     time.Time `json:"end_date" db:"end_date"`
	Capacity    int       `json:"capacity" db:"capacity"`
	HotelEmail  string    `json:"hotel_email,omitempty" db:"hotel_email"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
