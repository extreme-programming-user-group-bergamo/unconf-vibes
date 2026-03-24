package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	tuirooms "github.com/katurdays/unconf/internal/tui/rooms"
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

type fakeTeaModel struct{}

func (m fakeTeaModel) Init() tea.Cmd                           { return nil }
func (m fakeTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m fakeTeaModel) View() string                            { return "" }

type fakeSelectionModel struct {
	fakeTeaModel
	selection tuirooms.BookingSelection
	ok        bool
}

func (m fakeSelectionModel) BookingSelection() (tuirooms.BookingSelection, bool) {
	return m.selection, m.ok
}

type errWriter struct{}

func (errWriter) Write(_ []byte) (int, error) {
	return 0, io.ErrClosedPipe
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

func TestDefaultTerminalCapabilityChecker_NonTTYStdout(t *testing.T) {
	originalStdoutStat := terminalStdoutStat
	originalStdinStat := terminalStdinStat
	originalTermEnv := terminalEnv
	t.Cleanup(func() {
		terminalStdoutStat = originalStdoutStat
		terminalStdinStat = originalStdinStat
		terminalEnv = originalTermEnv
	})

	terminalStdoutStat = func() (os.FileMode, error) {
		return 0, nil
	}
	terminalStdinStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalEnv = func(_ string) string {
		return "xterm-256color"
	}

	assert.False(t, defaultTerminalCapabilityChecker{}.SupportsInteractiveUI())
}

func TestDefaultTerminalCapabilityChecker_NonTTYStdin(t *testing.T) {
	originalStdoutStat := terminalStdoutStat
	originalStdinStat := terminalStdinStat
	originalTermEnv := terminalEnv
	t.Cleanup(func() {
		terminalStdoutStat = originalStdoutStat
		terminalStdinStat = originalStdinStat
		terminalEnv = originalTermEnv
	})

	terminalStdoutStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalStdinStat = func() (os.FileMode, error) {
		return 0, nil
	}
	terminalEnv = func(_ string) string {
		return "xterm-256color"
	}

	assert.False(t, defaultTerminalCapabilityChecker{}.SupportsInteractiveUI())
}

func TestDefaultTerminalCapabilityChecker_TERMUnsetAllowed(t *testing.T) {
	originalStdoutStat := terminalStdoutStat
	originalStdinStat := terminalStdinStat
	originalTermEnv := terminalEnv
	t.Cleanup(func() {
		terminalStdoutStat = originalStdoutStat
		terminalStdinStat = originalStdinStat
		terminalEnv = originalTermEnv
	})

	terminalStdoutStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalStdinStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalEnv = func(_ string) string {
		return ""
	}

	assert.True(t, defaultTerminalCapabilityChecker{}.SupportsInteractiveUI())
}

func TestDefaultTerminalCapabilityChecker_TERMDumb(t *testing.T) {
	originalStdoutStat := terminalStdoutStat
	originalStdinStat := terminalStdinStat
	originalTermEnv := terminalEnv
	t.Cleanup(func() {
		terminalStdoutStat = originalStdoutStat
		terminalStdinStat = originalStdinStat
		terminalEnv = originalTermEnv
	})

	terminalStdoutStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalStdinStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalEnv = func(_ string) string {
		return "dumb"
	}

	assert.False(t, defaultTerminalCapabilityChecker{}.SupportsInteractiveUI())
}

func TestDefaultTerminalCapabilityChecker_TERMNormal(t *testing.T) {
	originalStdoutStat := terminalStdoutStat
	originalStdinStat := terminalStdinStat
	originalTermEnv := terminalEnv
	t.Cleanup(func() {
		terminalStdoutStat = originalStdoutStat
		terminalStdinStat = originalStdinStat
		terminalEnv = originalTermEnv
	})

	terminalStdoutStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalStdinStat = func() (os.FileMode, error) {
		return os.ModeCharDevice, nil
	}
	terminalEnv = func(_ string) string {
		return "xterm-256color"
	}

	assert.True(t, defaultTerminalCapabilityChecker{}.SupportsInteractiveUI())
}

func TestEmitBookingHandoffMessage_PrintsSelectionWhenAvailable(t *testing.T) {
	runModel := fakeSelectionModel{
		selection: tuirooms.BookingSelection{
			ConferenceSlug: "socrates-26",
			Room:           client.RoomResponse{ID: 9, RoomNumber: "304"},
		},
		ok: true,
	}

	var out bytes.Buffer
	err := emitBookingHandoffMessage(&out, runModel)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Booking flow handoff")
	assert.Contains(t, out.String(), "conference=socrates-26")
	assert.Contains(t, out.String(), "room=304")
}

func TestEmitBookingHandoffMessage_NoSelectionNoOutput(t *testing.T) {
	runModel := fakeSelectionModel{ok: false}
	var out bytes.Buffer

	err := emitBookingHandoffMessage(&out, runModel)
	require.NoError(t, err)
	assert.Empty(t, out.String())
}

func TestEmitBookingHandoffMessage_WriteFailure(t *testing.T) {
	runModel := fakeSelectionModel{
		selection: tuirooms.BookingSelection{
			ConferenceSlug: "socrates-26",
			Room:           client.RoomResponse{ID: 1, RoomNumber: "101"},
		},
		ok: true,
	}

	err := emitBookingHandoffMessage(errWriter{}, runModel)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write booking handoff message")
}
