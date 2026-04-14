package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// DashboardQuery contains optional organizer dashboard filters.
type DashboardQuery struct {
	RoomType           string
	BookingStatus      string
	HasSpecialRequests bool
	Search             string
}

// OrganizerDashboardRoomFillRateResponse represents occupancy summary by room type.
type OrganizerDashboardRoomFillRateResponse struct {
	RoomType      string  `json:"room_type"`
	Registrations int     `json:"registrations"`
	Capacity      int     `json:"capacity"`
	FillRatePct   float64 `json:"fill_rate_pct"`
}

// OrganizerDashboardAttendeeResponse represents one attendee row in dashboard payload.
type OrganizerDashboardAttendeeResponse struct {
	Name                     string `json:"name"`
	Email                    string `json:"email"`
	RoomNumber               string `json:"room_number,omitempty"`
	RoomType                 string `json:"room_type,omitempty"`
	BookingStatus            string `json:"booking_status"`
	DietaryAccessibilityNote string `json:"dietary_accessibility_note,omitempty"`
	PrivacySetting           string `json:"privacy_setting"`
}

// OrganizerDashboardResponse represents GET /conferences/{slug}/dashboard response.
type OrganizerDashboardResponse struct {
	ConferenceSlug      string                                   `json:"conference_slug"`
	TotalRegistrations  int                                      `json:"total_registrations"`
	Capacity            int                                      `json:"capacity"`
	CapacityUsagePct    float64                                  `json:"capacity_usage_pct"`
	RoomFillRates       []OrganizerDashboardRoomFillRateResponse `json:"room_fill_rates"`
	Attendees           []OrganizerDashboardAttendeeResponse     `json:"attendees"`
	AppliedRoomType     string                                   `json:"applied_room_type,omitempty"`
	AppliedBookingState string                                   `json:"applied_booking_status,omitempty"`
	AppliedSearch       string                                   `json:"applied_search,omitempty"`
}

// GetOrganizerDashboard fetches organizer dashboard data for a conference.
func (c *Client) GetOrganizerDashboard(
	ctx context.Context,
	accessToken string,
	slug string,
	query DashboardQuery,
) (*OrganizerDashboardResponse, error) {
	var result OrganizerDashboardResponse
	var errEnvelope apiErrorEnvelope

	request := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&result).
		SetError(&errEnvelope)

	if roomType := strings.TrimSpace(query.RoomType); roomType != "" {
		request.SetQueryParam("room_type", roomType)
	}
	if bookingStatus := strings.TrimSpace(query.BookingStatus); bookingStatus != "" {
		request.SetQueryParam("booking_status", bookingStatus)
	}
	if query.HasSpecialRequests {
		request.SetQueryParam("has_special_requests", "true")
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		request.SetQueryParam("q", search)
	}

	resp, err := request.Get("/conferences/" + url.PathEscape(slug) + "/dashboard")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizer dashboard: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusUnauthorized:
		return nil, fmt.Errorf("failed to fetch organizer dashboard: %w", ErrUnauthorized)
	case http.StatusForbidden:
		return nil, fmt.Errorf("failed to fetch organizer dashboard: %w", ErrOrganizerForbidden)
	case http.StatusNotFound:
		return nil, fmt.Errorf("failed to fetch organizer dashboard: %w", ErrConferenceNotFound)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to fetch organizer dashboard: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result.Attendees == nil {
		result.Attendees = []OrganizerDashboardAttendeeResponse{}
	}
	if result.RoomFillRates == nil {
		result.RoomFillRates = []OrganizerDashboardRoomFillRateResponse{}
	}

	return &result, nil
}
