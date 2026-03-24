package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBookClient struct {
	listRoomsFn     func(ctx context.Context, slug string) ([]client.RoomResponse, error)
	getMeFn         func(ctx context.Context) (*client.UserResponse, error)
	createBookingFn func(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error)
}

func (m *mockBookClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx, slug)
	}
	return []client.RoomResponse{}, nil
}

func (m *mockBookClient) GetMe(ctx context.Context) (*client.UserResponse, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx)
	}
	return &client.UserResponse{PrivacySetting: "public"}, nil
}

func (m *mockBookClient) CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
	if m.createBookingFn != nil {
		return m.createBookingFn(ctx, input)
	}
	return &client.BookingResponse{ID: 1, Status: "confirmed", PrivacySetting: input.PrivacySetting, Notes: input.Notes}, nil
}

type mockBookContextStore struct {
	getActiveConferenceFn func() (string, error)
}

func (m *mockBookContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

func executeBookCmd(t *testing.T, cmd *cobra.Command, args []string, stdin string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(stdin))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func executeBookCmdWithInput(t *testing.T, cmd *cobra.Command, args []string, input io.Reader) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(input)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

type errReader struct {
	err error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

func TestBookCmd_CreatesBookingWithProfilePrivacyAndNotes(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "socrates-26", nil
		},
	}

	var capturedInput client.CreateBookingRequest
	bookClient.listRoomsFn = func(_ context.Context, slug string) ([]client.RoomResponse, error) {
		require.Equal(t, "socrates-26", slug)
		return []client.RoomResponse{{ID: 42, ConferenceID: 7, RoomNumber: "101"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "private"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		capturedInput = input
		return &client.BookingResponse{
			ID:             9,
			Status:         "confirmed",
			PrivacySetting: input.PrivacySetting,
			Notes:          input.Notes,
		}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"101", "--notes", "vegetarian", "--yes"}, "")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, int64(42), capturedInput.RoomID)
	assert.Equal(t, int64(7), capturedInput.ConferenceID)
	assert.Equal(t, "private", capturedInput.PrivacySetting)
	assert.Equal(t, "vegetarian", capturedInput.Notes)
	assert.Contains(t, stdout, "Booking created successfully")
	assert.Contains(t, stdout, "Room:       101")
}

func TestBookCmd_PrivateFlagOverridesProfileDefault(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 10, ConferenceID: 8, RoomNumber: "202"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		assert.Equal(t, "private", input.PrivacySetting)
		return &client.BookingResponse{ID: 12, Status: "confirmed", PrivacySetting: input.PrivacySetting}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	_, _, err := executeBookCmd(t, cmd, []string{"202", "--private", "--yes"}, "")
	require.NoError(t, err)
}

func TestBookCmd_ConfirmationPromptCancelSkipsCreate(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 1, ConferenceID: 2, RoomNumber: "303"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, fmt.Errorf("should not be called")
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"303"}, "n\n")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Book room 303")
	assert.Contains(t, stdout, "Booking canceled")
}

func TestBookCmd_CreateBookingErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		roomNumber     string
		createErr      error
		expectedStderr string
		expectedErrIs  error
	}{
		{
			name:           "room full",
			roomNumber:     "404",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrRoomFull),
			expectedStderr: "Room 404 is full",
			expectedErrIs:  client.ErrRoomFull,
		},
		{
			name:           "already booked",
			roomNumber:     "505",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrAlreadyBooked),
			expectedStderr: "already have a booking",
			expectedErrIs:  client.ErrAlreadyBooked,
		},
		{
			name:           "create path room not found",
			roomNumber:     "606",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrRoomNotFound),
			expectedStderr: "Room 606 was not found in conference socrates-26",
			expectedErrIs:  client.ErrRoomNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookClient := &mockBookClient{}
			ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

			bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
				return []client.RoomResponse{{ID: 1, ConferenceID: 2, RoomNumber: tt.roomNumber}}, nil
			}
			bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
				return &client.UserResponse{PrivacySetting: "public"}, nil
			}
			bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
				return nil, tt.createErr
			}

			cmd := newBookCmd(bookClient, ctxStore)
			_, stderr, err := executeBookCmd(t, cmd, []string{tt.roomNumber, "--yes"}, "")

			require.Error(t, err)
			assert.Contains(t, stderr, tt.expectedStderr)
			if errors.Is(tt.createErr, client.ErrRoomNotFound) {
				assert.Contains(t, stderr, "Refresh room list and try again")
			}
			assert.Contains(t, err.Error(), "failed to create booking")
			assert.ErrorIs(t, err, tt.expectedErrIs)
		})
	}
}

func TestBookCmd_ConfirmationPromptAcceptsEOFWithInput(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 11, ConferenceID: 22, RoomNumber: "707"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		return &client.BookingResponse{ID: 1, Status: "confirmed", PrivacySetting: input.PrivacySetting}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"707"}, "yes")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Booking created successfully")
}

func TestBookCmd_ConfirmationPromptReadErrorReturnsFailure(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 13, ConferenceID: 24, RoomNumber: "808"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, fmt.Errorf("should not be called")
	}

	readErr := errors.New("stdin exploded")
	cmd := newBookCmd(bookClient, ctxStore)
	_, stderr, err := executeBookCmdWithInput(t, cmd, []string{"808"}, &errReader{err: readErr})

	require.Error(t, err)
	assert.Contains(t, stderr, "failed to confirm booking")
	assert.Contains(t, err.Error(), "failed to confirm booking")
	assert.ErrorIs(t, err, readErr)
}

func TestBookCmd_RoomNotFoundByArgument(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 7, ConferenceID: 8, RoomNumber: "101"}}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	_, _, err := executeBookCmd(t, cmd, []string{"999", "--yes"}, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "room \"999\" not found")
	assert.ErrorIs(t, err, client.ErrRoomNotFound)
}
