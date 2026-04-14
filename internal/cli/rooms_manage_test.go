package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomsAddCommand_UsesContextAndCreatesRoom(t *testing.T) {
	mockClient := &mockRoomsClient{
		createRoomFn: func(_ context.Context, slug string, input client.ManageRoomRequest) (*client.RoomResponse, error) {
			assert.Equal(t, "conf-53", slug)
			assert.Equal(t, "101", input.RoomNumber)
			return &client.RoomResponse{RoomNumber: "101"}, nil
		},
	}
	ctxStore := &mockRoomsContextStore{getActiveConferenceFn: func() (string, error) { return "conf-53", nil }}
	cmd := newRoomsAddCmd(mockClient, ctxStore)
	cmd.SetIn(bytes.NewBufferString("101\ndouble\n120\n2\n"))
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Room 101 created successfully")
}

func TestRoomsImportCommand_PartialFailuresReturnError(t *testing.T) {
	mockClient := &mockRoomsClient{
		createRoomFn: func(_ context.Context, _ string, input client.ManageRoomRequest) (*client.RoomResponse, error) {
			if input.RoomNumber == "102" {
				return nil, errors.New("duplicate")
			}
			return &client.RoomResponse{RoomNumber: input.RoomNumber}, nil
		},
	}
	ctxStore := &mockRoomsContextStore{getActiveConferenceFn: func() (string, error) { return "conf-53", nil }}

	tmp := t.TempDir() + "/rooms.csv"
	require.NoError(t, os.WriteFile(tmp, []byte("room_number,room_type,price_per_night,capacity\n101,double,100,2\n102,single,80,1\n"), 0o600))

	cmd := newRoomsImportCmd(mockClient, ctxStore)
	cmd.SetArgs([]string{tmp})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRoomsRemoveCommand_MapsBookingsError(t *testing.T) {
	mockClient := &mockRoomsClient{
		deleteRoomFn: func(_ context.Context, _, _ string) error { return client.ErrRoomHasBookings },
	}
	ctxStore := &mockRoomsContextStore{getActiveConferenceFn: func() (string, error) { return "conf-53", nil }}
	cmd := newRoomsRemoveCmd(mockClient, ctxStore)
	cmd.SetArgs([]string{"101"})
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)
	cmd.SetOut(&bytes.Buffer{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, errBuf.String(), "Cannot remove room")
}
