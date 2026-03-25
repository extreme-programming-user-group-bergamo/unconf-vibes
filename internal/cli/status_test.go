package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatusClient struct {
	bookings   []client.BookingResponse
	bookErr    error
	conference *client.ConferenceResponse
	confErr    error
}

func (m *mockStatusClient) ListBookings(_ context.Context) ([]client.BookingResponse, error) {
	if m.bookErr != nil {
		return nil, m.bookErr
	}

	return m.bookings, nil
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
			},
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
}

func TestStatusCmd_AllFlagRendersAcrossConferences(t *testing.T) {
	statusClient := &mockStatusClient{
		bookings: []client.BookingResponse{
			{ConferenceSlug: "conf-a", Conference: client.BookingConferenceResponse{Name: "Conf A"}, Room: client.BookingRoomResponse{RoomNumber: "101", RoomType: "single", PricePerNight: 100}},
			{ConferenceSlug: "conf-b", Conference: client.BookingConferenceResponse{Name: "Conf B"}, Room: client.BookingRoomResponse{RoomNumber: "202", RoomType: "double", PricePerNight: 220}},
		},
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

func TestStatusCmd_ContextReadFailure(t *testing.T) {
	cmd := newStatusCmd(&mockStatusClient{}, &mockStatusContextStore{err: errors.New("boom")})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read conference context")
}
