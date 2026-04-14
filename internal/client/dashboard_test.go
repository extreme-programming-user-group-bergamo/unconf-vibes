package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOrganizerDashboard_SuccessWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conferences/socrates-26/dashboard", r.URL.Path)
		assert.Equal(t, "double", r.URL.Query().Get("room_type"))
		assert.Equal(t, "confirmed", r.URL.Query().Get("booking_status"))
		assert.Equal(t, "true", r.URL.Query().Get("has_special_requests"))
		assert.Equal(t, "ali", r.URL.Query().Get("q"))
		assert.Equal(t, "Bearer valid-access-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(OrganizerDashboardResponse{
			ConferenceSlug:     "socrates-26",
			TotalRegistrations: 2,
			RoomFillRates: []OrganizerDashboardRoomFillRateResponse{
				{RoomType: "double", Registrations: 2, Capacity: 4, FillRatePct: 50},
			},
			Attendees: []OrganizerDashboardAttendeeResponse{
				{Name: "Alice", Email: "alice@test.dev", PrivacySetting: "private", BookingStatus: "confirmed"},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	resp, err := c.GetOrganizerDashboard(context.Background(), "valid-access-token", "socrates-26", DashboardQuery{
		RoomType:           "double",
		BookingStatus:      "confirmed",
		HasSpecialRequests: true,
		Search:             "ali",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "socrates-26", resp.ConferenceSlug)
	require.Len(t, resp.Attendees, 1)
	assert.Equal(t, "Alice", resp.Attendees[0].Name)
}

func TestGetOrganizerDashboard_MapsErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    error
	}{
		{name: "unauthorized", statusCode: http.StatusUnauthorized, wantErr: ErrUnauthorized},
		{name: "forbidden", statusCode: http.StatusForbidden, wantErr: ErrOrganizerForbidden},
		{name: "not found", statusCode: http.StatusNotFound, wantErr: ErrConferenceNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.statusCode)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{"code": "x", "message": "x"},
				})
			}))
			defer srv.Close()

			c := NewClient(srv.URL)
			resp, err := c.GetOrganizerDashboard(context.Background(), "token", "slug", DashboardQuery{})
			require.Error(t, err)
			assert.Nil(t, resp)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
