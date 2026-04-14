package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	tuirooms "github.com/katurdays/unconf/internal/tui/rooms"
	tuiwizard "github.com/katurdays/unconf/internal/tui/wizard"
	"github.com/spf13/cobra"
)

// RoomsClient defines the room listing API contract.
type RoomsClient interface {
	ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error)
	CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error)
	ListBookings(ctx context.Context) ([]client.BookingResponse, error)
	CreateRoommateRequest(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error)
	CreateRoom(ctx context.Context, slug string, input client.ManageRoomRequest) (*client.RoomResponse, error)
	UpdateRoom(ctx context.Context, slug string, roomNumber string, input client.ManageRoomRequest) (*client.RoomResponse, error)
	DeleteRoom(ctx context.Context, slug string, roomNumber string) error
}

// RoomsContextStore defines the interface for reading active conference context.
type RoomsContextStore interface {
	GetActiveConference() (string, error)
}

type terminalCapabilityChecker interface {
	SupportsInteractiveUI() bool
}

type defaultTerminalCapabilityChecker struct{}

var launchRoomsExplorer = func(model tea.Model) (tea.Model, error) {
	return tea.NewProgram(model).Run()
}

var launchBookingWizard = func(model tea.Model) (tea.Model, error) {
	return tea.NewProgram(model).Run()
}

var terminalStdoutStat = func() (os.FileMode, error) {
	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return 0, err
	}

	return stdoutInfo.Mode(), nil
}

var terminalStdinStat = func() (os.FileMode, error) {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return 0, err
	}

	return stdinInfo.Mode(), nil
}

var terminalEnv = func(key string) string {
	return os.Getenv(key)
}

func (defaultTerminalCapabilityChecker) SupportsInteractiveUI() bool {
	stdoutMode, err := terminalStdoutStat()
	if err != nil {
		return false
	}
	stdinMode, err := terminalStdinStat()
	if err != nil {
		return false
	}

	if stdoutMode&os.ModeCharDevice == 0 {
		return false
	}
	if stdinMode&os.ModeCharDevice == 0 {
		return false
	}

	term := strings.TrimSpace(strings.ToLower(terminalEnv("TERM")))
	return term != "dumb"
}

func newRoomsCmd(roomsClient RoomsClient, ctxStore RoomsContextStore) *cobra.Command {
	return newRoomsCmdWithChecker(roomsClient, ctxStore, defaultTerminalCapabilityChecker{})
}

func newRoomsCmdWithChecker(roomsClient RoomsClient, ctxStore RoomsContextStore, checker terminalCapabilityChecker) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rooms [conference-slug]",
		Short: "Browse conference rooms",
		Long:  "Launches the room explorer TUI for a conference. If no slug is provided, uses the active conference context.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveRoomsSlug(cmd, ctxStore, args)
			if err != nil {
				return err
			}
			if slug == "" {
				return nil
			}

			return runRooms(cmd, roomsClient, checker, slug)
		},
	}

	cmd.AddCommand(
		newRoomsAddCmd(roomsClient, ctxStore),
		newRoomsImportCmd(roomsClient, ctxStore),
		newRoomsEditCmd(roomsClient, ctxStore),
		newRoomsRemoveCmd(roomsClient, ctxStore),
	)

	return cmd
}

func resolveRoomsSlug(cmd *cobra.Command, ctxStore RoomsContextStore, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}

	if slug == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Usage: unconf rooms <slug> or set context with 'unconf checkout <slug>'")
		return "", nil
	}

	return slug, nil
}

func runRooms(cmd *cobra.Command, roomsClient RoomsClient, checker terminalCapabilityChecker, slug string) error {
	if !checker.SupportsInteractiveUI() {
		slog.Warn("rooms: interactive terminal unsupported, falling back to plain output", "slug", slug)
		return runRoomsFallback(cmd, roomsClient, slug)
	}

	slog.Info("rooms: starting interactive explorer", "slug", slug)

	ownRoomID, _ := resolveOwnRoomID(cmd.Context(), roomsClient, slug)
	model := tuirooms.NewModel(cmd.Context(), slug, roomsClient.ListRooms, ownRoomID, common.NewStyles())
	runModel, err := launchRoomsExplorer(model)
	if err != nil {
		return fmt.Errorf("failed to launch rooms explorer: %w", err)
	}

	inviteSelection, inviteSelected := extractInviteSelection(runModel)
	if inviteSelected {
		if err := runRoomsInviteFlow(cmd, roomsClient, slug, inviteSelection); err != nil {
			return err
		}
		return nil
	}

	selection, selected := extractBookingSelection(runModel)
	if !selected {
		return nil
	}

	if err := runBookingWizardFlow(cmd, roomsClient, selection); err != nil {
		return err
	}

	return nil
}

func extractBookingSelection(runModel tea.Model) (tuirooms.BookingSelection, bool) {
	selectionProvider, ok := runModel.(interface {
		BookingSelection() (tuirooms.BookingSelection, bool)
	})
	if !ok {
		return tuirooms.BookingSelection{}, false
	}

	return selectionProvider.BookingSelection()
}

func extractInviteSelection(runModel tea.Model) (tuirooms.InviteSelection, bool) {
	inviteProvider, ok := runModel.(interface {
		InviteSelection() (tuirooms.InviteSelection, bool)
	})
	if !ok {
		return tuirooms.InviteSelection{}, false
	}

	return inviteProvider.InviteSelection()
}

func runBookingWizardFlow(cmd *cobra.Command, roomsClient RoomsClient, selection tuirooms.BookingSelection) error {
	wizardModel := tuiwizard.NewModel(
		cmd.Context(),
		tuiwizard.RoomSelection{
			ConferenceSlug: selection.ConferenceSlug,
			Room:           selection.Room,
		},
		roomsClient.CreateBooking,
		common.NewStyles(),
	)

	if _, err := launchBookingWizard(wizardModel); err != nil {
		return fmt.Errorf("failed to launch booking wizard for room %s in conference %s: %w", selection.Room.RoomNumber, selection.ConferenceSlug, err)
	}

	return nil
}

func runRoomsInviteFlow(cmd *cobra.Command, roomsClient RoomsClient, conferenceSlug string, selection tuirooms.InviteSelection) error {
	if selection.Room.SpotsAvailable < 1 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "You can only invite when your room has available spots.")
		return nil
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), "Invite GitHub username: ")
	usernameInput, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) || strings.TrimSpace(usernameInput) == "" {
			return fmt.Errorf("failed to read invite username: %w", err)
		}
	}

	username := normalizeInviteUsername(usernameInput)
	if username == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Invite canceled.")
		return nil
	}

	if _, err := createInviteForRoom(cmd.Context(), roomsClient, selection.Room.ID, username); err != nil {
		handleInviteError(cmd, conferenceSlug, username, err)
		return fmt.Errorf("failed to send invite to @%s from room view: %w", username, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Request sent to @%s\n", username)
	return nil
}

func resolveOwnRoomID(ctx context.Context, roomsClient RoomsClient, conferenceSlug string) (int64, error) {
	bookings, err := roomsClient.ListBookings(ctx)
	if err != nil {
		return 0, err
	}

	filtered := filterBookingsByConference(bookings, 0, conferenceSlug)
	if len(filtered) == 0 {
		return 0, nil
	}

	return filtered[0].RoomID, nil
}

func runRoomsFallback(cmd *cobra.Command, roomsClient RoomsClient, slug string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	rooms, err := roomsClient.ListRooms(cmd.Context(), slug)
	if err != nil {
		if errors.Is(err, client.ErrConferenceNotFound) {
			_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", slug)
			return fmt.Errorf("failed to list rooms: %w", err)
		}
		return fmt.Errorf("failed to list rooms: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Interactive room explorer unavailable in this terminal.\n")
	_, _ = fmt.Fprintf(out, "Showing plain room list for conference %q.\n\n", slug)

	if len(rooms) == 0 {
		_, _ = fmt.Fprintln(out, "No rooms found.")
		_, _ = fmt.Fprintln(out, "Tip: use an interactive terminal (TTY with TERM set) to launch full TUI mode.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ROOM\tTYPE\tAVAILABILITY\tPRICE")
	for _, room := range rooms {
		availability := fmt.Sprintf("%d/%d", room.SpotsAvailable, room.Capacity)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\n", room.RoomNumber, room.RoomType, availability, room.PricePerNight)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to render room fallback table: %w", flushErr)
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Tip: use an interactive terminal (TTY with TERM set) to launch full TUI mode.")

	return nil
}
