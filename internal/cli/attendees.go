package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"text/tabwriter"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// AttendeesClient defines API operations needed for attendee listing.
type AttendeesClient interface {
	ListAttendees(ctx context.Context, slug string) (*client.AttendeeListResponse, error)
}

// AttendeesContextStore defines the interface for reading active conference context.
type AttendeesContextStore interface {
	GetActiveConference() (string, error)
}

func newAttendeesCmd(attendeesClient AttendeesClient, ctxStore AttendeesContextStore) *cobra.Command {
	return &cobra.Command{
		Use:   "attendees [conference-slug]",
		Short: "List conference attendees",
		Long:  "Displays public attendees for a conference. If no slug is provided, uses the active conference context.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveAttendeesSlug(cmd, ctxStore, args)
			if err != nil {
				return err
			}
			if slug == "" {
				return nil
			}

			return runAttendees(cmd, attendeesClient, slug)
		},
	}
}

func resolveAttendeesSlug(cmd *cobra.Command, ctxStore AttendeesContextStore, args []string) (string, error) {
	if len(args) == 1 {
		if slug := strings.TrimSpace(args[0]); slug != "" {
			return slug, nil
		}
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}
	if slug == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active conference context. Run 'unconf checkout <slug>' or use 'unconf attendees <slug>'.")
		return "", nil
	}

	return slug, nil
}

func runAttendees(cmd *cobra.Command, attendeesClient AttendeesClient, conferenceSlug string) error {
	slog.Info("attendees: fetching attendee list", "conference_slug", conferenceSlug)

	result, err := attendeesClient.ListAttendees(cmd.Context(), conferenceSlug)
	if err != nil {
		return handleAttendeesFetchError(cmd, conferenceSlug, err)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(out, "Attendees for conference %q:\n", conferenceSlug)

	if len(result.Attendees) > 0 {
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tROOM")
		for i := range result.Attendees {
			_, _ = fmt.Fprintf(w, "%s\t%s\n", attendeeDisplayName(result.Attendees[i]), attendeeRoomLabel(result.Attendees[i]))
		}
		_ = w.Flush()
	} else {
		_, _ = fmt.Fprintln(out, "No public attendees yet.")
	}

	if result.PrivateAttendeesCount > 0 {
		_, _ = fmt.Fprintf(out, "+ %d private attendees\n", result.PrivateAttendeesCount)
	}

	return nil
}

func handleAttendeesFetchError(cmd *cobra.Command, conferenceSlug string, err error) error {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return fmt.Errorf("failed to list attendees for %q: %w", conferenceSlug, auth.ErrNotAuthenticated)
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
		return fmt.Errorf("failed to list attendees for %q: %w", conferenceSlug, client.ErrSessionExpired)
	case errors.Is(err, client.ErrUnauthorized):
		_, _ = fmt.Fprintln(errOut, "You are not authorized. Please run 'unconf login' and try again.")
		return fmt.Errorf("failed to list attendees for %q: %w", conferenceSlug, client.ErrUnauthorized)
	case errors.Is(err, client.ErrConferenceNotFound):
		_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", conferenceSlug)
		return fmt.Errorf("failed to list attendees for %q: %w", conferenceSlug, client.ErrConferenceNotFound)
	default:
		_, _ = fmt.Fprintln(errOut, "Could not fetch attendee list from the API. Please try again.")
		return fmt.Errorf("failed to list attendees for %q: %w", conferenceSlug, err)
	}
}

func attendeeDisplayName(attendee client.AttendeeProjectionResponse) string {
	name := strings.TrimSpace(attendee.DisplayName)
	if name != "" {
		return name
	}

	username := strings.TrimSpace(attendee.GitHubUsername)
	if username != "" {
		return "@" + username
	}

	return "Attendee"
}

func attendeeRoomLabel(attendee client.AttendeeProjectionResponse) string {
	if attendee.Room == nil {
		return "not booked yet"
	}

	roomNumber := strings.TrimSpace(attendee.Room.RoomNumber)
	roomType := strings.TrimSpace(attendee.Room.RoomType)
	if roomNumber == "" {
		return "not booked yet"
	}
	if roomType == "" {
		return roomNumber
	}

	return fmt.Sprintf("%s (%s)", roomNumber, roomType)
}
