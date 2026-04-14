package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ManageRoomRequest is the payload for organizer room create/update endpoints.
type ManageRoomRequest struct {
	RoomNumber    string  `json:"room_number"`
	RoomType      string  `json:"room_type"`
	PricePerNight float64 `json:"price_per_night"`
	Capacity      int     `json:"capacity"`
}

// RoomOccupantResponse represents a room occupant with privacy-aware fields.
type RoomOccupantResponse struct {
	UserID      *int64 `json:"user_id,omitempty"`
	DisplayName string `json:"display_name"`
}

// RoomResponse represents a room returned by GET /conferences/{slug}/rooms.
type RoomResponse struct {
	ID             int64                  `json:"id"`
	ConferenceID   int64                  `json:"conference_id"`
	RoomNumber     string                 `json:"room_number"`
	RoomType       string                 `json:"room_type"`
	PricePerNight  float64                `json:"price_per_night"`
	Capacity       int                    `json:"capacity"`
	SpotsTaken     int                    `json:"spots_taken"`
	SpotsAvailable int                    `json:"spots_available"`
	Occupants      []RoomOccupantResponse `json:"occupants"`
}

// ListRooms fetches room availability for a conference via GET /conferences/{slug}/rooms.
func (c *Client) ListRooms(ctx context.Context, slug string) ([]RoomResponse, error) {
	var result []RoomResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/conferences/" + url.PathEscape(slug) + "/rooms")
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrConferenceNotFound
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list rooms: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result == nil {
		result = []RoomResponse{}
	}

	return result, nil
}

// CreateRoom creates a room via POST /conferences/{slug}/rooms.
func (c *Client) CreateRoom(ctx context.Context, accessToken string, slug string, input ManageRoomRequest) (*RoomResponse, error) {
	var result RoomResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetBody(input).
		SetResult(&result).
		SetError(&errEnvelope).
		Post("/conferences/" + url.PathEscape(slug) + "/rooms")
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to create room: %w", ErrUnauthorized)
	}
	if resp.StatusCode() == http.StatusForbidden {
		return nil, fmt.Errorf("failed to create room: %w", ErrOrganizerForbidden)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, fmt.Errorf("failed to create room: %w", ErrConferenceNotFound)
	}
	if resp.StatusCode() == http.StatusConflict {
		return nil, fmt.Errorf("failed to create room: %w", ErrRoomExists)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to create room: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}

// UpdateRoom updates a room via PUT /conferences/{slug}/rooms/{number}.
func (c *Client) UpdateRoom(ctx context.Context, accessToken string, slug string, roomNumber string, input ManageRoomRequest) (*RoomResponse, error) {
	var result RoomResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetBody(input).
		SetResult(&result).
		SetError(&errEnvelope).
		Put("/conferences/" + url.PathEscape(slug) + "/rooms/" + url.PathEscape(roomNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to update room: %w", ErrUnauthorized)
	}
	if resp.StatusCode() == http.StatusForbidden {
		return nil, fmt.Errorf("failed to update room: %w", ErrOrganizerForbidden)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return nil, fmt.Errorf("failed to update room: %w", ErrRoomNotFound)
	}
	if resp.StatusCode() == http.StatusConflict {
		return nil, fmt.Errorf("failed to update room: %w", ErrRoomExists)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to update room: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}

// DeleteRoom removes a room via DELETE /conferences/{slug}/rooms/{number}.
func (c *Client) DeleteRoom(ctx context.Context, accessToken string, slug string, roomNumber string) error {
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetError(&errEnvelope).
		Delete("/conferences/" + url.PathEscape(slug) + "/rooms/" + url.PathEscape(roomNumber))
	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return fmt.Errorf("failed to delete room: %w", ErrUnauthorized)
	}
	if resp.StatusCode() == http.StatusForbidden {
		return fmt.Errorf("failed to delete room: %w", ErrOrganizerForbidden)
	}
	if resp.StatusCode() == http.StatusNotFound {
		return fmt.Errorf("failed to delete room: %w", ErrRoomNotFound)
	}
	if resp.StatusCode() == http.StatusConflict {
		return fmt.Errorf("failed to delete room: %w", ErrRoomHasBookings)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to delete room: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return nil
}
