package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

const (
	privacyPublic  = "public"
	privacyPrivate = "private"
)

// BookClient defines API operations needed for direct booking.
type BookClient interface {
	ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error)
	GetMe(ctx context.Context) (*client.UserResponse, error)
	CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error)
}

// BookContextStore defines the interface for reading active conference context.
type BookContextStore interface {
	GetActiveConference() (string, error)
}

func newBookCmd(bookClient BookClient, ctxStore BookContextStore) *cobra.Command {
	var notes string
	var forcePrivate bool
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "book <room_number>",
		Short: "Book a room directly",
		Long:  "Creates a booking directly for the active conference context without launching the TUI booking wizard.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			roomNumber := strings.TrimSpace(args[0])
			if roomNumber == "" {
				return fmt.Errorf("invalid argument: room number cannot be empty")
			}

			conferenceSlug, err := ctxStore.GetActiveConference()
			if err != nil {
				return fmt.Errorf("failed to read conference context: %w", err)
			}
			if conferenceSlug == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No active conference context. Run 'unconf checkout <slug>' before booking.")
				return nil
			}

			room, err := findRoomByNumber(cmd.Context(), bookClient, conferenceSlug, roomNumber)
			if err != nil {
				return err
			}

			privacySetting, err := resolvePrivacySetting(cmd.Context(), bookClient, cmd.Flags().Changed("private"), forcePrivate)
			if err != nil {
				if errors.Is(err, auth.ErrNotAuthenticated) {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "You are not logged in. Run 'unconf login' to authenticate.")
					return fmt.Errorf("failed to resolve privacy setting: %w", err)
				}
				if errors.Is(err, client.ErrSessionExpired) {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Your session has expired. Please run 'unconf login' to re-authenticate.")
					return fmt.Errorf("failed to resolve privacy setting: %w", err)
				}
				return fmt.Errorf("failed to resolve privacy setting: %w", err)
			}

			trimmedNotes := strings.TrimSpace(notes)
			if !skipConfirm {
				confirmed, confirmErr := promptForBookingConfirmation(cmd, conferenceSlug, room.RoomNumber, privacySetting, trimmedNotes)
				if confirmErr != nil {
					return fmt.Errorf("failed to confirm booking: %w", confirmErr)
				}
				if !confirmed {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Booking canceled.")
					return nil
				}
			}

			bookingInput := client.CreateBookingRequest{
				RoomID:         room.ID,
				ConferenceID:   room.ConferenceID,
				PrivacySetting: privacySetting,
				Notes:          trimmedNotes,
			}

			booking, createErr := bookClient.CreateBooking(cmd.Context(), bookingInput)
			if createErr != nil {
				handleCreateBookingError(cmd, conferenceSlug, room.RoomNumber, createErr)
				return fmt.Errorf("failed to create booking for room %s in conference %s: %w", room.RoomNumber, conferenceSlug, createErr)
			}

			renderBookingSuccess(cmd, conferenceSlug, room.RoomNumber, booking)
			return nil
		},
	}

	cmd.Flags().BoolVar(&forcePrivate, "private", false, "Book privately")
	cmd.Flags().StringVar(&notes, "notes", "", "Hotel notes to include with the booking")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip booking confirmation prompt")

	return cmd
}

func findRoomByNumber(ctx context.Context, bookClient BookClient, conferenceSlug string, roomNumber string) (*client.RoomResponse, error) {
	rooms, err := bookClient.ListRooms(ctx, conferenceSlug)
	if err != nil {
		if errors.Is(err, client.ErrConferenceNotFound) {
			return nil, fmt.Errorf("conference %q not found: %w", conferenceSlug, err)
		}
		return nil, fmt.Errorf("failed to list rooms for conference %s: %w", conferenceSlug, err)
	}

	for i := range rooms {
		if rooms[i].RoomNumber == roomNumber {
			return &rooms[i], nil
		}
	}

	return nil, fmt.Errorf("room %q not found in conference %q: %w", roomNumber, conferenceSlug, client.ErrRoomNotFound)
}

func resolvePrivacySetting(ctx context.Context, bookClient BookClient, privateFlagChanged bool, forcePrivate bool) (string, error) {
	if privateFlagChanged && forcePrivate {
		return privacyPrivate, nil
	}

	user, err := bookClient.GetMe(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load user profile: %w", err)
	}

	privacy := strings.TrimSpace(strings.ToLower(user.PrivacySetting))
	if privacy == privacyPrivate {
		return privacyPrivate, nil
	}

	return privacyPublic, nil
}

func promptForBookingConfirmation(cmd *cobra.Command, conferenceSlug string, roomNumber string, privacySetting string, notes string) (bool, error) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Book room %s in conference %s with privacy %s", roomNumber, conferenceSlug, privacySetting)
	if notes != "" {
		_, _ = fmt.Fprintf(out, " and notes %q", notes)
	}
	_, _ = fmt.Fprint(out, "? [y/N]: ")

	input, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) || strings.TrimSpace(input) == "" {
			return false, fmt.Errorf("failed to read confirmation input: %w", err)
		}
	}

	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func handleCreateBookingError(cmd *cobra.Command, conferenceSlug string, roomNumber string, err error) {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, client.ErrRoomFull):
		_, _ = fmt.Fprintf(errOut, "Room %s is full in conference %s. Try a different room.\n", roomNumber, conferenceSlug)
	case errors.Is(err, client.ErrAlreadyBooked):
		_, _ = fmt.Fprintln(errOut, "You already have a booking for this conference. Cancel it before creating a new one.")
	case errors.Is(err, client.ErrRoomNotFound):
		_, _ = fmt.Fprintf(errOut, "Room %s was not found in conference %s. Refresh room list and try again.\n", roomNumber, conferenceSlug)
	default:
		_, _ = fmt.Fprintf(errOut, "Could not create booking for room %s in conference %s. Please try again.\n", roomNumber, conferenceSlug)
	}
}

func renderBookingSuccess(cmd *cobra.Command, conferenceSlug string, roomNumber string, booking *client.BookingResponse) {
	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Booking created successfully.")
	_, _ = fmt.Fprintf(out, "  Conference: %s\n", conferenceSlug)
	_, _ = fmt.Fprintf(out, "  Room:       %s\n", roomNumber)
	_, _ = fmt.Fprintf(out, "  Booking ID: %d\n", booking.ID)
	_, _ = fmt.Fprintf(out, "  Status:     %s\n", booking.Status)
	_, _ = fmt.Fprintf(out, "  Privacy:    %s\n", booking.PrivacySetting)
	if strings.TrimSpace(booking.Notes) != "" {
		_, _ = fmt.Fprintf(out, "  Notes:      %s\n", booking.Notes)
	}
}
