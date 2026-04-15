package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCancelClient struct {
	listFn   func(context.Context) ([]client.BookingResponse, error)
	cancelFn func(context.Context, int64) (*client.BookingResponse, error)
}

func (m *mockCancelClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockCancelClient) CancelBooking(ctx context.Context, bookingID int64) (*client.BookingResponse, error) {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, bookingID)
	}
	return &client.BookingResponse{ID: bookingID, Status: "cancelled"}, nil
}

type mockCancelContextStore struct {
	getActiveConferenceFn func() (string, error)
}

func (m *mockCancelContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

func executeCancelCmd(t *testing.T, cmd *cobra.Command, args []string, stdin string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(stdin))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCancelCmd_ConfirmsAndCancels(t *testing.T) {
	cancelCalled := false
	cmd := newCancelCmd(&mockCancelClient{
		listFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{ID: 9, ConferenceSlug: "socrates-26", Room: client.BookingRoomResponse{RoomNumber: "204"}}}, nil
		},
		cancelFn: func(_ context.Context, bookingID int64) (*client.BookingResponse, error) {
			cancelCalled = true
			assert.Equal(t, int64(9), bookingID)
			return &client.BookingResponse{ID: bookingID, Status: "cancelled"}, nil
		},
	}, &mockCancelContextStore{
		getActiveConferenceFn: func() (string, error) { return "socrates-26", nil },
	})

	stdout, stderr, err := executeCancelCmd(t, cmd, []string{}, "yes\n")
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.True(t, cancelCalled)
	assert.Contains(t, stdout, "Cancel your booking in Room 204?")
	assert.Contains(t, stdout, "This action is irreversible")
	assert.Contains(t, stdout, "Type 'yes' or 'y' to continue [y/N]:")
	assert.Contains(t, stdout, "Booking cancelled successfully")
	assert.Contains(t, stdout, "roommates have been notified")
}

func TestCancelCmd_DeclinedConfirmationAborts(t *testing.T) {
	cancelCalled := false
	cmd := newCancelCmd(&mockCancelClient{
		listFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{ID: 9, ConferenceSlug: "socrates-26", Room: client.BookingRoomResponse{RoomNumber: "204"}}}, nil
		},
		cancelFn: func(_ context.Context, _ int64) (*client.BookingResponse, error) {
			cancelCalled = true
			return &client.BookingResponse{Status: "cancelled"}, nil
		},
	}, &mockCancelContextStore{
		getActiveConferenceFn: func() (string, error) { return "socrates-26", nil },
	})

	stdout, stderr, err := executeCancelCmd(t, cmd, []string{}, "no\n")
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.False(t, cancelCalled)
	assert.Contains(t, stdout, "Cancellation aborted.")
}

func TestCancelCmd_NoActiveConferenceContext(t *testing.T) {
	cmd := newCancelCmd(&mockCancelClient{}, &mockCancelContextStore{
		getActiveConferenceFn: func() (string, error) { return "", nil },
	})

	stdout, stderr, err := executeCancelCmd(t, cmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, stdout)
	assert.Contains(t, stderr, "No active conference context")
}

func TestCancelCmd_YesFlagSkipsConfirmation(t *testing.T) {
	cancelCalled := false
	cmd := newCancelCmd(&mockCancelClient{
		listFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{{ID: 3, ConferenceSlug: "socrates-26", Room: client.BookingRoomResponse{RoomNumber: "310"}}}, nil
		},
		cancelFn: func(_ context.Context, bookingID int64) (*client.BookingResponse, error) {
			cancelCalled = true
			return &client.BookingResponse{ID: bookingID, Status: "cancelled"}, nil
		},
	}, &mockCancelContextStore{
		getActiveConferenceFn: func() (string, error) { return "socrates-26", nil },
	})

	stdout, stderr, err := executeCancelCmd(t, cmd, []string{"--yes"}, "")
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.True(t, cancelCalled)
	assert.NotContains(t, stdout, "Cancel your booking in Room")
}

func TestCancelCmd_NoActiveBooking(t *testing.T) {
	cmd := newCancelCmd(&mockCancelClient{
		listFn: func(_ context.Context) ([]client.BookingResponse, error) {
			return []client.BookingResponse{}, nil
		},
	}, &mockCancelContextStore{
		getActiveConferenceFn: func() (string, error) { return "socrates-26", nil },
	})

	stdout, stderr, err := executeCancelCmd(t, cmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "No active booking found for conference")
}

func TestCancelCmd_ErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStderr string
	}{
		{name: "not-authenticated", err: auth.ErrNotAuthenticated, wantStderr: "not logged in"},
		{name: "session-expired", err: client.ErrSessionExpired, wantStderr: "session has expired"},
		{name: "unauthorized", err: client.ErrUnauthorized, wantStderr: "not authorized"},
		{name: "not-found", err: client.ErrBookingNotFound, wantStderr: "No active booking found to cancel"},
		{name: "forbidden", err: client.ErrBookingForbidden, wantStderr: "cancel your own booking"},
		{name: "generic", err: errors.New("boom"), wantStderr: "Could not cancel booking"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newCancelCmd(&mockCancelClient{
				listFn: func(_ context.Context) ([]client.BookingResponse, error) {
					return []client.BookingResponse{{ID: 1, ConferenceSlug: "socrates-26", Room: client.BookingRoomResponse{RoomNumber: "204"}}}, nil
				},
				cancelFn: func(_ context.Context, _ int64) (*client.BookingResponse, error) {
					return nil, tc.err
				},
			}, &mockCancelContextStore{
				getActiveConferenceFn: func() (string, error) { return "socrates-26", nil },
			})

			_, stderr, err := executeCancelCmd(t, cmd, []string{"--yes"}, "")
			require.Error(t, err)
			assert.Contains(t, stderr, tc.wantStderr)
		})
	}
}
