package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ExportConferenceBookingsCSV fetches organizer CSV export for a conference.
func (c *Client) ExportConferenceBookingsCSV(
	ctx context.Context,
	accessToken string,
	slug string,
	includeCancelled bool,
) ([]byte, error) {
	var errEnvelope apiErrorEnvelope

	request := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetError(&errEnvelope)

	if includeCancelled {
		request.SetQueryParam("include_cancelled", "true")
	}

	resp, err := request.Get("/conferences/" + url.PathEscape(slug) + "/export")
	if err != nil {
		return nil, fmt.Errorf("failed to export conference bookings: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("failed to export conference bookings: %w", ErrUnauthorized)
	case http.StatusForbidden:
		return nil, fmt.Errorf("failed to export conference bookings: %w", ErrOrganizerForbidden)
	case http.StatusNotFound:
		return nil, fmt.Errorf("failed to export conference bookings: %w", ErrConferenceNotFound)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to export conference bookings: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return resp.Body(), nil
}
