package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

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
