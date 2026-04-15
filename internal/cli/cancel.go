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

type CancelClient interface {
	ListBookings(ctx context.Context) ([]client.BookingResponse, error)
	CancelBooking(ctx context.Context, bookingID int64) (*client.BookingResponse, error)
}

type CancelContextStore interface {
	GetActiveConference() (string, error)
}

func newCancelCmd(cancelClient CancelClient, ctxStore CancelContextStore) *cobra.Command {
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel your current booking",
		RunE: func(cmd *cobra.Command, _ []string) error {
			conferenceSlug, err := ctxStore.GetActiveConference()
			if err != nil {
				return fmt.Errorf("failed to read conference context: %w", err)
			}
			if conferenceSlug == "" {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No active conference context. Run 'unconf checkout <slug>' before cancelling.")
				return nil
			}

			bookings, err := cancelClient.ListBookings(cmd.Context())
			if err != nil {
				handleCancelError(cmd, err)
				return fmt.Errorf("failed to resolve booking for conference %s: %w", conferenceSlug, err)
			}

			filtered := filterBookingsByConference(bookings, 0, conferenceSlug)
			if len(filtered) == 0 {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No active booking found for conference %q.\n", conferenceSlug)
				return nil
			}

			booking := filtered[0]
			roomNumber := bookingRoomNumber(booking)
			if !skipConfirm {
				confirmed, confirmErr := promptCancelConfirmation(cmd, roomNumber)
				if confirmErr != nil {
					return fmt.Errorf("failed to confirm cancellation: %w", confirmErr)
				}
				if !confirmed {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancellation aborted.")
					return nil
				}
			}

			_, err = cancelClient.CancelBooking(cmd.Context(), booking.ID)
			if err != nil {
				handleCancelError(cmd, err)
				return fmt.Errorf("failed to cancel booking %d: %w", booking.ID, err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Booking cancelled successfully.")
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Your roommates have been notified and will keep their room assignment.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip cancellation confirmation prompt")
	return cmd
}

func promptCancelConfirmation(cmd *cobra.Command, roomNumber string) (bool, error) {
	_, _ = fmt.Fprintf(
		cmd.OutOrStdout(),
		"Cancel your booking in Room %s? This action is irreversible. Type 'yes' or 'y' to continue [y/N]: ",
		roomNumber,
	)

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

func handleCancelError(cmd *cobra.Command, err error) {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
	case errors.Is(err, client.ErrUnauthorized):
		_, _ = fmt.Fprintln(errOut, "You are not authorized. Please run 'unconf login' and try again.")
	case errors.Is(err, client.ErrBookingNotFound):
		_, _ = fmt.Fprintln(errOut, "No active booking found to cancel.")
	case errors.Is(err, client.ErrBookingForbidden):
		_, _ = fmt.Fprintln(errOut, "You can only cancel your own booking.")
	default:
		_, _ = fmt.Fprintln(errOut, "Could not cancel booking. Please try again.")
	}
}
