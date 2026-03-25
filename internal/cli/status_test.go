package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatusClient struct {
	bookings   []client.BookingResponse
	bookErr    error
	requests   []client.RoommateRequestResponse
	requestErr error
	conference *client.ConferenceResponse
	confErr    error
}

func (m *mockStatusClient) ListBookings(_ context.Context) ([]client.BookingResponse, error) {
	if m.bookErr != nil {
		return nil, m.bookErr
	}

	return m.bookings, nil
}

func (m *mockStatusClient) ListRoommateRequests(_ context.Context) ([]client.RoommateRequestResponse, error) {
	if m.requestErr != nil {
		return nil, m.requestErr
	}

	return m.requests, nil
}

func (m *mockStatusClient) GetConference(_ context.Context, _ string) (*client.ConferenceResponse, error) {
	if m.confErr != nil {
		return nil, m.confErr
	}

	if m.conference == nil {
		return &client.ConferenceResponse{}, nil
	}

	return m.conference, nil
}

type mockStatusContextStore struct {
	activeConference string
	err              error
}

func (m *mockStatusContextStore) GetActiveConference() (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.activeConference, nil
}

func TestStatusCmd_UsesActiveConferenceAndRendersBookingProjection(t *testing.T) {
	statusClient := &mockStatusClient{
		conference: &client.ConferenceResponse{ID: 2, Slug: "socrates-2026"},
		bookings: []client.BookingResponse{
			{
				ID:             11,
				RoomID:         7,
				ConferenceID:   2,
				Status:         "confirmed",
				PrivacySetting: "public",
				Room: client.BookingRoomResponse{
					RoomNumber:    "101",
					RoomType:      "double",
					PricePerNight: 189.5,
				},
				Conference: client.BookingConferenceResponse{
					Slug:      "socrates-2026",
					Name:      "SoCraTes 2026",
					StartDate: "2026-09-10",
					EndDate:   "2026-09-12",
				},
				Roommates: []client.BookingRoommateResponse{
					{DisplayName: "Alice", PrivacySetting: "public"},
					{DisplayName: "Bob", PrivacySetting: "private"},
				},
			},
		},
		requests: []client.RoommateRequestResponse{
			{ConferenceID: 2, Status: "pending", Direction: "incoming"},
			{ConferenceID: 2, Status: "pending", Direction: "outgoing"},
		},
	}

	cmd := newStatusCmd(statusClient, &mockStatusContextStore{activeConference: "socrates-2026"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Booking status for conference \"socrates-2026\"")
	assert.Contains(t, output, "Room:       101 (double)")
	assert.Contains(t, output, "Price:      $189.50/night")
	assert.Contains(t, output, "Dates:      2026-09-10 - 2026-09-12")
	assert.Contains(t, output, "Privacy:    public")
	assert.Contains(t, output, "Roommates:")
	assert.Contains(t, output, "- Alice")
	assert.Contains(t, output, "- Private attendee")
	assert.Contains(t, output, "Pending roommate requests (Epic 4 placeholder)")
	assert.Contains(t, output, "Incoming: 1")
	assert.Contains(t, output, "Outgoing: 1")
}

func TestStatusCmd_AllFlagRendersAcrossConferences(t *testing.T) {
	statusClient := &mockStatusClient{
		bookings: []client.BookingResponse{
			{ConferenceSlug: "conf-a", Conference: client.BookingConferenceResponse{Name: "Conf A"}, Room: client.BookingRoomResponse{RoomNumber: "101", RoomType: "single", PricePerNight: 100}},
			{ConferenceSlug: "conf-b", Conference: client.BookingConferenceResponse{Name: "Conf B"}, Room: client.BookingRoomResponse{RoomNumber: "202", RoomType: "double", PricePerNight: 220}},
		},
		requests: []client.RoommateRequestResponse{{ConferenceSlug: "conf-a", Status: "pending", Direction: "incoming"}},
	}

	cmd := newStatusCmd(statusClient, &mockStatusContextStore{})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Booking status across all conferences")
	assert.Contains(t, output, "Conference: Conf A (conf-a)")
	assert.Contains(t, output, "Conference: Conf B (conf-b)")
}

func TestStatusCmd_NoActiveConferenceMessage(t *testing.T) {
	cmd := newStatusCmd(&mockStatusClient{}, &mockStatusContextStore{})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No active conference context")
}

func TestStatusCmd_NoBookingForActiveConferenceMessage(t *testing.T) {
	statusClient := &mockStatusClient{
		conference: &client.ConferenceResponse{ID: 99, Slug: "socrates-2026"},
		bookings: []client.BookingResponse{
			{ConferenceID: 42, ConferenceSlug: "other-conf"},
		},
	}

	cmd := newStatusCmd(statusClient, &mockStatusContextStore{activeConference: "socrates-2026"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No booking found for active conference")
	assert.Contains(t, stdout.String(), "unconf book <room_number>")
}

func TestStatusCmd_ContextReadFailure(t *testing.T) {
	cmd := newStatusCmd(&mockStatusClient{}, &mockStatusContextStore{err: errors.New("boom")})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read conference context")
}

func TestStatusCmd_AuthErrorMapping_NotAuthenticated(t *testing.T) {
	cmd := newStatusCmd(&mockStatusClient{bookErr: auth.ErrNotAuthenticated}, &mockStatusContextStore{activeConference: "conf-1"})
	cmd.SetOut(&bytes.Buffer{})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
	assert.Contains(t, stderr.String(), "not logged in")
}

func TestStatusCmd_AuthErrorMapping_SessionExpired(t *testing.T) {
	cmd := newStatusCmd(&mockStatusClient{bookErr: client.ErrSessionExpired}, &mockStatusContextStore{activeConference: "conf-1"})
	cmd.SetOut(&bytes.Buffer{})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrSessionExpired)
	assert.Contains(t, stderr.String(), "session has expired")
}
