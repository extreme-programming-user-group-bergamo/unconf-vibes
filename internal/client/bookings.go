package client

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// BookingRoomResponse represents room data projected into booking payloads.
type BookingRoomResponse struct {
	ID            int64   `json:"id"`
	RoomNumber    string  `json:"room_number"`
	RoomType      string  `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
}

// BookingConferenceResponse represents conference data projected into booking payloads.
type BookingConferenceResponse struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// BookingRoommateResponse represents roommate details projected in booking payloads.
type BookingRoommateResponse struct {
	DisplayName    string `json:"display_name"`
	PrivacySetting string `json:"privacy_setting"`
}

// CreateBookingRequest represents the payload for POST /bookings.
type CreateBookingRequest struct {
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	PrivacySetting string `json:"privacy_setting"`
	Notes          string `json:"notes,omitempty"`
}

// BookingResponse represents a booking returned by booking API endpoints.
type BookingResponse struct {
	ID             int64                     `json:"id"`
	RoomID         int64                     `json:"room_id"`
	ConferenceID   int64                     `json:"conference_id"`
	ConferenceSlug string                    `json:"conference_slug,omitempty"`
	Status         string                    `json:"status"`
	PrivacySetting string                    `json:"privacy_setting"`
	Notes          string                    `json:"notes,omitempty"`
	CreatedAt      string                    `json:"created_at"`
	ConfirmedAt    string                    `json:"confirmed_at,omitempty"`
	CancelledAt    string                    `json:"cancelled_at,omitempty"`
	Room           BookingRoomResponse       `json:"room"`
	Conference     BookingConferenceResponse `json:"conference"`
	Roommates      []BookingRoommateResponse `json:"roommates"`
}

// ListBookings fetches the authenticated user's bookings via GET /bookings.
func (c *Client) ListBookings(ctx context.Context, accessToken string) ([]BookingResponse, error) {
	var result []BookingResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/bookings")
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to list bookings: %w", ErrUnauthorized)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list bookings: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result == nil {
		result = []BookingResponse{}
	}

	return result, nil
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

// CancelBooking cancels a booking via DELETE /bookings/{id}.
func (c *Client) CancelBooking(ctx context.Context, accessToken string, bookingID int64) (*BookingResponse, error) {
	var result BookingResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&result).
		SetError(&errEnvelope).
		Delete("/bookings/" + strconv.FormatInt(bookingID, 10))
	if err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to cancel booking: %w", ErrUnauthorized)
	}

	if resp.StatusCode() == http.StatusForbidden {
		return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingForbidden)
	}

	if resp.StatusCode() == http.StatusNotFound && errEnvelope.Error.Code == "not_found" {
		return nil, fmt.Errorf("failed to cancel booking: %w", ErrBookingNotFound)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to cancel booking: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}
