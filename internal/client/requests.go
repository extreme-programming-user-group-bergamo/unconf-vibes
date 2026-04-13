package client

import (
	"context"
	"fmt"
	"net/http"
)

// CreateRoommateRequestRequest represents the payload for POST /requests.
type CreateRoommateRequestRequest struct {
	TargetUsername string `json:"target_username,omitempty"`
	TargetID       int64  `json:"target_id,omitempty"`
	RoomID         int64  `json:"room_id"`
}

// RoommateRequestResponse represents a roommate request returned by GET /requests.
type RoommateRequestResponse struct {
	ID             int64  `json:"id"`
	RequesterID    int64  `json:"requester_id"`
	TargetID       int64  `json:"target_id"`
	RoomID         int64  `json:"room_id"`
	ConferenceID   int64  `json:"conference_id"`
	ConferenceSlug string `json:"conference_slug,omitempty"`
	Status         string `json:"status"`
	Direction      string `json:"direction,omitempty"`
	RequesterName  string `json:"requester_name,omitempty"`
	TargetName     string `json:"target_name,omitempty"`
	CreatedAt      string `json:"created_at"`
	RespondedAt    string `json:"responded_at,omitempty"`
}

// CreateRoommateRequest creates a roommate request via POST /requests.
func (c *Client) CreateRoommateRequest(ctx context.Context, accessToken string, input CreateRoommateRequestRequest) (*RoommateRequestResponse, error) {
	var result RoommateRequestResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetBody(input).
		SetResult(&result).
		SetError(&errEnvelope).
		Post("/requests")
	if err != nil {
		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to create roommate request: %w", ErrUnauthorized)
	}

	if resp.StatusCode() == http.StatusNotFound {
		if errEnvelope.Error.Code == "target_not_found" {
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrTargetUserNotFound)
		}
	}

	if resp.StatusCode() == http.StatusConflict {
		switch errEnvelope.Error.Code {
		case "target_has_booking":
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrTargetAlreadyBooked)
		case "duplicate_request":
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrRequestPending)
		case "room_full":
			return nil, fmt.Errorf("failed to create roommate request: %w", ErrRoomFull)
		}
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to create roommate request: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}

// ListRoommateRequests fetches roommate requests via GET /requests.
func (c *Client) ListRoommateRequests(ctx context.Context, accessToken string) ([]RoommateRequestResponse, error) {
	var result []RoommateRequestResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/requests")
	if err != nil {
		return nil, fmt.Errorf("failed to list roommate requests: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to list roommate requests: %w", ErrUnauthorized)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list roommate requests: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result == nil {
		result = []RoommateRequestResponse{}
	}

	return result, nil
}
