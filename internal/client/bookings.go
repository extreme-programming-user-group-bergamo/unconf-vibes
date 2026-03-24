package client

import (
	"context"
	"fmt"
	"net/http"
)

// CreateBookingRequest represents the payload for POST /bookings.
type CreateBookingRequest struct {
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	PrivacySetting string `json:"privacy_setting"`
	Notes          string `json:"notes,omitempty"`
}

// BookingResponse represents a booking returned by booking API endpoints.
type BookingResponse struct {
	ID             int64  `json:"id"`
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	Status         string `json:"status"`
	PrivacySetting string `json:"privacy_setting"`
	Notes          string `json:"notes,omitempty"`
	CreatedAt      string `json:"created_at"`
	ConfirmedAt    string `json:"confirmed_at,omitempty"`
	CancelledAt    string `json:"cancelled_at,omitempty"`
}

// CreateBooking creates a booking via POST /bookings.
func (c *Client) CreateBooking(ctx context.Context, accessToken string, input CreateBookingRequest) (*BookingResponse, error) {
	var result BookingResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetBody(input).
		SetResult(&result).
		SetError(&errEnvelope).
		Post("/bookings")
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to create booking: %w", ErrUnauthorized)
	}

	if resp.StatusCode() == http.StatusConflict {
		switch errEnvelope.Error.Code {
		case "room_full":
			return nil, fmt.Errorf("failed to create booking: %w", ErrRoomFull)
		case "already_booked":
			return nil, fmt.Errorf("failed to create booking: %w", ErrAlreadyBooked)
		}
	}

	if resp.StatusCode() == http.StatusNotFound {
		switch errEnvelope.Error.Code {
		case "not_found":
			return nil, fmt.Errorf("failed to create booking: %w", ErrRoomNotFound)
		}
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create booking: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}
