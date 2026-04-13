package models

import "time"

type RoommateRequestStatus string

const (
	RoommateRequestStatusPending  RoommateRequestStatus = "pending"
	RoommateRequestStatusAccepted RoommateRequestStatus = "accepted"
	RoommateRequestStatusDeclined RoommateRequestStatus = "declined"
)

type RoommateRequest struct {
	ID          int64                 `json:"id" db:"id"`
	RequesterID int64                 `json:"requester_id" db:"requester_id"`
	TargetID    int64                 `json:"target_id" db:"target_id"`
	RoomID      int64                 `json:"room_id" db:"room_id"`
	Status      RoommateRequestStatus `json:"status" db:"status"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
}
