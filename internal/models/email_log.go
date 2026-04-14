package models

import "time"

type EmailLogStatus string

const (
	EmailLogStatusSent   EmailLogStatus = "sent"
	EmailLogStatusFailed EmailLogStatus = "failed"
)

type EmailLog struct {
	ID           int64          `json:"id" db:"id"`
	BookingID    int64          `json:"booking_id" db:"booking_id"`
	EmailType    string         `json:"email_type" db:"email_type"`
	Recipient    string         `json:"recipient" db:"recipient"`
	Subject      string         `json:"subject" db:"subject"`
	Status       EmailLogStatus `json:"status" db:"status"`
	Attempt      int            `json:"attempt" db:"attempt"`
	ErrorDetails string         `json:"error_details,omitempty" db:"error_details"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
}
