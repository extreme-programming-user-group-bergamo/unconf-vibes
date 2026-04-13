package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockInviteClient struct {
	listBookingsFn func(ctx context.Context) ([]client.BookingResponse, error)
	createFn       func(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error)
}

func (m *mockInviteClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	if m.listBookingsFn != nil {
		return m.listBookingsFn(ctx)
	}
	return []client.BookingResponse{}, nil
}

func (m *mockInviteClient) CreateRoommateRequest(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return &client.RoommateRequestResponse{ID: 1, Status: "pending"}, nil
}

type mockInviteContextStore struct {
	activeConference string
	err              error
}

func (m *mockInviteContextStore) GetActiveConference() (string, error) {
	return m.activeConference, m.err
}

func executeInviteCmd(t *testing.T, cmd *cobra.Command, args []string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestInviteCmd_SendsRequest(t *testing.T) {
	inviteClient := &mockInviteClient{
		listBookingsFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{RoomID: 42, Conference: client.BookingConferenceResponse{Slug: "socrates-26"}}}, nil
		},
		createFn: func(_ context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
			assert.Equal(t, int64(42), input.RoomID)
			assert.Equal(t, "alice", input.TargetUsername)
			return &client.RoommateRequestResponse{ID: 4, Status: "pending"}, nil
		},
	}

	cmd := newInviteCmd(inviteClient, &mockInviteContextStore{activeConference: "socrates-26"})
	stdout, stderr, err := executeInviteCmd(t, cmd, []string{"@alice"})
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Request sent to @alice")
}

func TestInviteCmd_TargetHasBookingShowsError(t *testing.T) {
	inviteClient := &mockInviteClient{
		listBookingsFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{RoomID: 42, Conference: client.BookingConferenceResponse{Slug: "socrates-26"}}}, nil
		},
		createFn: func(_ context.Context, _ client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
			return nil, client.ErrTargetAlreadyBooked
		},
	}

	cmd := newInviteCmd(inviteClient, &mockInviteContextStore{activeConference: "socrates-26"})
	_, stderr, err := executeInviteCmd(t, cmd, []string{"bob"})
	require.Error(t, err)
	assert.Contains(t, stderr, "already has a booking")
	assert.ErrorIs(t, err, client.ErrTargetAlreadyBooked)
}

func TestInviteCmd_PendingRequestShowsError(t *testing.T) {
	inviteClient := &mockInviteClient{
		listBookingsFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{RoomID: 42, Conference: client.BookingConferenceResponse{Slug: "socrates-26"}}}, nil
		},
		createFn: func(_ context.Context, _ client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
			return nil, client.ErrRequestPending
		},
	}

	cmd := newInviteCmd(inviteClient, &mockInviteContextStore{activeConference: "socrates-26"})
	_, stderr, err := executeInviteCmd(t, cmd, []string{"bob"})
	require.Error(t, err)
	assert.Contains(t, stderr, "pending roommate request")
	assert.ErrorIs(t, err, client.ErrRequestPending)
}

func TestInviteCmd_RequiresOwnBooking(t *testing.T) {
	inviteClient := &mockInviteClient{
		listBookingsFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{}, nil
		},
	}

	cmd := newInviteCmd(inviteClient, &mockInviteContextStore{activeConference: "socrates-26"})
	_, stderr, err := executeInviteCmd(t, cmd, []string{"bob"})
	require.Error(t, err)
	assert.Contains(t, stderr, "Could not send invite")
}

func TestInviteCmd_NoConferenceContext(t *testing.T) {
	cmd := newInviteCmd(&mockInviteClient{}, &mockInviteContextStore{})
	stdout, stderr, err := executeInviteCmd(t, cmd, []string{"bob"})
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "No active conference context")
}

func TestNormalizeInviteUsername(t *testing.T) {
	assert.Equal(t, "alice", normalizeInviteUsername(" @alice "))
	assert.Equal(t, "", normalizeInviteUsername("   "))
}

func TestResolveOwnConferenceBooking_PropagatesError(t *testing.T) {
	_, err := resolveOwnConferenceBooking(context.Background(), &mockInviteClient{
		listBookingsFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return nil, errors.New("boom")
		},
	}, "socrates-26")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
