package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDashboardClient struct {
	getFn func(ctx context.Context, slug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error)
}

func (m *mockDashboardClient) GetOrganizerDashboard(ctx context.Context, slug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
	if m.getFn != nil {
		return m.getFn(ctx, slug, query)
	}
	return &client.OrganizerDashboardResponse{}, nil
}

type mockDashboardContextStore struct {
	activeConference string
	err              error
}

func (m *mockDashboardContextStore) GetActiveConference() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.activeConference, nil
}

type fakeDashboardModel struct{}

func (m fakeDashboardModel) Init() tea.Cmd                           { return nil }
func (m fakeDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m fakeDashboardModel) View() string                            { return "" }

func TestDashboardCmd_FallbackWhenTerminalUnsupported(t *testing.T) {
	cmd := newDashboardCmdWithChecker(&mockDashboardClient{
		getFn: func(_ context.Context, slug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			assert.Equal(t, "double", query.RoomType)
			return &client.OrganizerDashboardResponse{
				ConferenceSlug:     slug,
				TotalRegistrations: 1,
				Capacity:           2,
				CapacityUsagePct:   50,
				RoomFillRates: []client.OrganizerDashboardRoomFillRateResponse{
					{RoomType: "double", Registrations: 1, Capacity: 2, FillRatePct: 50},
				},
				Attendees: []client.OrganizerDashboardAttendeeResponse{
					{Name: "Alice", Email: "alice@test.dev", RoomNumber: "101", RoomType: "double", BookingStatus: "confirmed", PrivacySetting: "private"},
				},
			}, nil
		},
	}, &mockDashboardContextStore{}, mockTerminalChecker{supports: false})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"socrates-26", "--room-type", "double"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Interactive organizer dashboard unavailable")
	assert.Contains(t, stdout.String(), "Total registrations: 1")
	assert.Contains(t, stdout.String(), "Alice")
	assert.Empty(t, stderr.String())
}

func TestDashboardCmd_InteractiveLaunchesTUI(t *testing.T) {
	originalLauncher := launchDashboardProgram
	t.Cleanup(func() { launchDashboardProgram = originalLauncher })

	launched := false
	launchDashboardProgram = func(model tea.Model) (tea.Model, error) {
		launched = true
		return fakeDashboardModel{}, nil
	}

	cmd := newDashboardCmdWithChecker(&mockDashboardClient{
		getFn: func(_ context.Context, _ string, _ client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
			return &client.OrganizerDashboardResponse{}, nil
		},
	}, &mockDashboardContextStore{}, mockTerminalChecker{supports: true})

	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, launched)
}

func TestDashboardCmd_ErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantError  error
		wantStderr string
	}{
		{name: "not authenticated", err: auth.ErrNotAuthenticated, wantError: auth.ErrNotAuthenticated, wantStderr: "not logged in"},
		{name: "session expired", err: client.ErrSessionExpired, wantError: client.ErrSessionExpired, wantStderr: "session has expired"},
		{name: "forbidden", err: client.ErrOrganizerForbidden, wantError: client.ErrOrganizerForbidden, wantStderr: "Organizer permissions required"},
		{name: "conference not found", err: client.ErrConferenceNotFound, wantError: client.ErrConferenceNotFound, wantStderr: "not found"},
		{name: "generic", err: errors.New("timeout"), wantStderr: "Could not fetch organizer dashboard"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newDashboardCmdWithChecker(&mockDashboardClient{
				getFn: func(_ context.Context, _ string, _ client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
					return nil, tc.err
				},
			}, &mockDashboardContextStore{}, mockTerminalChecker{supports: false})

			cmd.SetOut(&bytes.Buffer{})
			var stderr bytes.Buffer
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{"socrates-26"})

			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, stderr.String(), tc.wantStderr)
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			}
		})
	}
}
