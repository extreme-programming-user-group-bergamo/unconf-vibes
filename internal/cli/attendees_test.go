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

type mockAttendeesClient struct {
	result *client.AttendeeListResponse
	err    error
	slug   string
}

func (m *mockAttendeesClient) ListAttendees(_ context.Context, slug string) (*client.AttendeeListResponse, error) {
	m.slug = slug
	if m.err != nil {
		return nil, m.err
	}
	if m.result == nil {
		return &client.AttendeeListResponse{}, nil
	}

	return m.result, nil
}

type mockAttendeesContextStore struct {
	activeConference string
	err              error
}

func (m *mockAttendeesContextStore) GetActiveConference() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.activeConference, nil
}

func TestAttendeesCmd_UsesActiveConferenceAndRendersTable(t *testing.T) {
	mockClient := &mockAttendeesClient{
		result: &client.AttendeeListResponse{
			Attendees: []client.AttendeeProjectionResponse{
				{
					DisplayName:    "Alice",
					GitHubUsername: "alice",
					Room: &client.AttendeeRoomProjectionResponse{
						RoomNumber: "101",
						RoomType:   "double",
					},
				},
				{
					DisplayName: "Bob",
				},
			},
			PrivateAttendeesCount: 5,
		},
	}

	cmd := newAttendeesCmd(mockClient, &mockAttendeesContextStore{activeConference: "socrates-26"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Equal(t, "socrates-26", mockClient.slug)
	assert.Contains(t, output, "Attendees for conference \"socrates-26\"")
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "ROOM")
	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "101 (double)")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "not booked yet")
	assert.Contains(t, output, "+ 5 private attendees")
}

func TestAttendeesCmd_ExplicitSlugOverridesContext(t *testing.T) {
	mockClient := &mockAttendeesClient{}
	cmd := newAttendeesCmd(mockClient, &mockAttendeesContextStore{activeConference: "ignored-conf"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"manual-conf"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "manual-conf", mockClient.slug)
}

func TestAttendeesCmd_WhitespaceSlugFallsBackToContext(t *testing.T) {
	mockClient := &mockAttendeesClient{}
	cmd := newAttendeesCmd(mockClient, &mockAttendeesContextStore{activeConference: "ctx-conf"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"   "})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "ctx-conf", mockClient.slug)
}

func TestAttendeesCmd_NoContextMessage(t *testing.T) {
	cmd := newAttendeesCmd(&mockAttendeesClient{}, &mockAttendeesContextStore{})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No active conference context")
}

func TestAttendeesCmd_ContextReadFailure(t *testing.T) {
	cmd := newAttendeesCmd(&mockAttendeesClient{}, &mockAttendeesContextStore{err: errors.New("boom")})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read conference context")
}

func TestAttendeesCmd_ErrorMappings(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		wantErrorIs     error
		wantErrContains string
		wantStderr      string
	}{
		{
			name:            "not authenticated",
			err:             auth.ErrNotAuthenticated,
			wantErrorIs:     auth.ErrNotAuthenticated,
			wantErrContains: "failed to list attendees",
			wantStderr:      "not logged in",
		},
		{
			name:            "session expired",
			err:             client.ErrSessionExpired,
			wantErrorIs:     client.ErrSessionExpired,
			wantErrContains: "failed to list attendees",
			wantStderr:      "session has expired",
		},
		{
			name:            "unauthorized",
			err:             client.ErrUnauthorized,
			wantErrorIs:     client.ErrUnauthorized,
			wantErrContains: "failed to list attendees",
			wantStderr:      "not authorized",
		},
		{
			name:            "conference not found",
			err:             client.ErrConferenceNotFound,
			wantErrorIs:     client.ErrConferenceNotFound,
			wantErrContains: "failed to list attendees",
			wantStderr:      "not found",
		},
		{
			name:            "generic fallback",
			err:             errors.New("transport timeout"),
			wantErrContains: "failed to list attendees",
			wantStderr:      "Could not fetch attendee list from the API",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newAttendeesCmd(
				&mockAttendeesClient{err: tc.err},
				&mockAttendeesContextStore{activeConference: "conf-1"},
			)
			cmd.SetOut(&bytes.Buffer{})
			var stderr bytes.Buffer
			cmd.SetErr(&stderr)
			cmd.SetArgs([]string{})

			err := cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
			assert.Contains(t, stderr.String(), tc.wantStderr)

			if tc.wantErrorIs != nil {
				assert.ErrorIs(t, err, tc.wantErrorIs)
			}
		})
	}
}
