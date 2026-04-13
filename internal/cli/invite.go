package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type InviteClient interface {
	ListBookings(ctx context.Context) ([]client.BookingResponse, error)
	CreateRoommateRequest(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error)
}

type InviteContextStore interface {
	GetActiveConference() (string, error)
}

func newInviteCmd(inviteClient InviteClient, ctxStore InviteContextStore) *cobra.Command {
	return &cobra.Command{
		Use:   "invite <username>",
		Short: "Send a roommate request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			username := normalizeInviteUsername(args[0])
			if username == "" {
				return fmt.Errorf("invalid argument: username cannot be empty")
			}

			conferenceSlug, err := ctxStore.GetActiveConference()
			if err != nil {
				return fmt.Errorf("failed to read conference context: %w", err)
			}
			if conferenceSlug == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No active conference context. Run 'unconf checkout <slug>' before inviting.")
				return nil
			}

			booking, err := resolveOwnConferenceBooking(cmd.Context(), inviteClient, conferenceSlug)
			if err != nil {
				handleInviteError(cmd, conferenceSlug, username, err)
				return fmt.Errorf("failed to send invite to @%s: %w", username, err)
			}

			if _, err := createInviteForRoom(cmd.Context(), inviteClient, booking.RoomID, username); err != nil {
				handleInviteError(cmd, conferenceSlug, username, err)
				return fmt.Errorf("failed to send invite to @%s: %w", username, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Request sent to @%s\n", username)
			return nil
		},
	}
}

func normalizeInviteUsername(raw string) string {
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "@"))
}

func resolveOwnConferenceBooking(ctx context.Context, inviteClient InviteClient, conferenceSlug string) (*client.BookingResponse, error) {
	bookings, err := inviteClient.ListBookings(ctx)
	if err != nil {
		return nil, err
	}

	filtered := filterBookingsByConference(bookings, 0, conferenceSlug)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no booking found for conference %q", conferenceSlug)
	}

	return &filtered[0], nil
}

func createInviteForRoom(ctx context.Context, inviteClient InviteClient, roomID int64, username string) (*client.RoommateRequestResponse, error) {
	return inviteClient.CreateRoommateRequest(ctx, client.CreateRoommateRequestRequest{
		TargetUsername: username,
		RoomID:         roomID,
	})
}

func handleInviteError(cmd *cobra.Command, conferenceSlug string, username string, err error) {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
	case errors.Is(err, client.ErrUnauthorized):
		_, _ = fmt.Fprintln(errOut, "You are not authorized. Please run 'unconf login' and try again.")
	case errors.Is(err, client.ErrTargetUserNotFound):
		_, _ = fmt.Fprintf(errOut, "User @%s was not found.\n", username)
	case errors.Is(err, client.ErrTargetAlreadyBooked):
		_, _ = fmt.Fprintf(errOut, "User @%s already has a booking for conference %s.\n", username, conferenceSlug)
	case errors.Is(err, client.ErrRequestPending):
		_, _ = fmt.Fprintf(errOut, "A pending roommate request to @%s already exists.\n", username)
	case errors.Is(err, client.ErrRoomFull):
		_, _ = fmt.Fprintln(errOut, "You can only invite when your room has available spots.")
	default:
		_, _ = fmt.Fprintf(errOut, "Could not send invite to @%s. Please try again.\n", username)
	}
}
