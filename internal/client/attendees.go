package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AttendeeRoomProjectionResponse represents room details for an attendee row.
type AttendeeRoomProjectionResponse struct {
	RoomNumber string `json:"room_number"`
	RoomType   string `json:"room_type"`
}

// AttendeeProjectionResponse represents an attendee returned by GET /conferences/{slug}/attendees.
type AttendeeProjectionResponse struct {
	DisplayName    string                          `json:"display_name"`
	GitHubUsername string                          `json:"github_username,omitempty"`
	Room           *AttendeeRoomProjectionResponse `json:"room,omitempty"`
}

// AttendeeListResponse represents the attendee list payload.
type AttendeeListResponse struct {
	Attendees             []AttendeeProjectionResponse `json:"attendees"`
	PrivateAttendeesCount int                          `json:"private_attendees_count"`
}

// ListAttendees fetches attendees for a conference via GET /conferences/{slug}/attendees.
func (c *Client) ListAttendees(ctx context.Context, accessToken string, slug string) (*AttendeeListResponse, error) {
	var result AttendeeListResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/conferences/" + url.PathEscape(slug) + "/attendees")
	if err != nil {
		return nil, fmt.Errorf("failed to list attendees: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to list attendees: %w", ErrUnauthorized)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrConferenceNotFound
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list attendees: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result.Attendees == nil {
		result.Attendees = []AttendeeProjectionResponse{}
	}

	return &result, nil
}
