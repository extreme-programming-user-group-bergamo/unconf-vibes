package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoomsClient struct {
	listRoomsFn func(ctx context.Context, slug string) ([]client.RoomResponse, error)
}

func (m *mockRoomsClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx, slug)
	}
	return []client.RoomResponse{}, nil
}

type mockRoomsContextStore struct {
	getActiveConferenceFn func() (string, error)
}

func (m *mockRoomsContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

type mockTerminalChecker struct {
	supports bool
}

func (m mockTerminalChecker) SupportsInteractiveUI() bool {
	return m.supports
}

func TestRoomsCmd_FallbackWhenTerminalUnsupported(t *testing.T) {
	roomsClient := &mockRoomsClient{
		listRoomsFn: func(_ context.Context, slug string) ([]client.RoomResponse, error) {
			assert.Equal(t, "socrates-26", slug)
			return []client.RoomResponse{{
				RoomNumber:     "101",
				RoomType:       "double",
				SpotsAvailable: 1,
				Capacity:       2,
				PricePerNight:  120,
			}}, nil
		},
	}

	cmd := newRoomsCmdWithChecker(roomsClient, &mockRoomsContextStore{}, mockTerminalChecker{supports: false})
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, stdout.String(), "Interactive room explorer unavailable")
	assert.Contains(t, stdout.String(), "ROOM")
	assert.Contains(t, stdout.String(), "101")
	assert.Contains(t, stdout.String(), "Tip: use an interactive terminal")
	assert.Empty(t, stderr.String())
}

func TestRoomsCmd_UsesContextWhenSlugOmitted(t *testing.T) {
	roomsClient := &mockRoomsClient{
		listRoomsFn: func(_ context.Context, slug string) ([]client.RoomResponse, error) {
			assert.Equal(t, "from-context", slug)
			return []client.RoomResponse{}, nil
		},
	}
	ctxStore := &mockRoomsContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "from-context", nil
		},
	}

	cmd := newRoomsCmdWithChecker(roomsClient, ctxStore, mockTerminalChecker{supports: false})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
}

func TestRoomsCmd_NoSlugAndNoContext(t *testing.T) {
	cmd := newRoomsCmdWithChecker(&mockRoomsClient{}, &mockRoomsContextStore{}, mockTerminalChecker{supports: false})
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "No conference specified")
}

func TestRoomsCmd_ConferenceNotFoundInFallback(t *testing.T) {
	roomsClient := &mockRoomsClient{
		listRoomsFn: func(_ context.Context, _ string) ([]client.RoomResponse, error) {
			return nil, client.ErrConferenceNotFound
		},
	}

	cmd := newRoomsCmdWithChecker(roomsClient, &mockRoomsContextStore{}, mockTerminalChecker{supports: false})
	cmd.SetOut(&bytes.Buffer{})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"missing"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list rooms")
	assert.Contains(t, stderr.String(), "not found")
}

func TestRoomsCmd_FallbackPassesThroughGenericErrors(t *testing.T) {
	roomsClient := &mockRoomsClient{
		listRoomsFn: func(_ context.Context, _ string) ([]client.RoomResponse, error) {
			return nil, fmt.Errorf("network timeout")
		},
	}

	cmd := newRoomsCmdWithChecker(roomsClient, &mockRoomsContextStore{}, mockTerminalChecker{supports: false})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
}
