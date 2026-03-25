package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// StatusClient defines the interface used by the status command.
type StatusClient interface {
	ListBookings(ctx context.Context) ([]client.BookingResponse, error)
	ListRoommateRequests(ctx context.Context) ([]client.RoommateRequestResponse, error)
	GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error)
}

// StatusContextStore defines the interface for reading active conference context.
type StatusContextStore interface {
	GetActiveConference() (string, error)
}

func newStatusCmd(statusClient StatusClient, ctxStore StatusContextStore) *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show booking status",
		Long:  "Displays your booking status for the active conference context. Use --all to include bookings across all conferences.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd, statusClient, ctxStore, showAll)
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show bookings across all conferences")

	return cmd
}

func runStatus(cmd *cobra.Command, statusClient StatusClient, ctxStore StatusContextStore, showAll bool) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()

	activeConference, err := resolveStatusScope(ctxStore, showAll)
	if err != nil {
		return err
	}

	if !showAll && activeConference == "" {
		_, _ = fmt.Fprintln(out, "No active conference context. Run 'unconf checkout <slug>' or use 'unconf status --all'.")
		return nil
	}

	bookings, err := statusClient.ListBookings(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch bookings: %w", err)
	}

	requests, err := statusClient.ListRoommateRequests(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch roommate requests: %w", err)
	}

	if showAll {
		renderStatusAllConferences(out, bookings, requests)
		return nil
	}

	conference, confErr := statusClient.GetConference(ctx, activeConference)
	if confErr != nil {
		return fmt.Errorf("failed to resolve conference context %q: %w", activeConference, confErr)
	}

	filtered := filterBookingsByConference(bookings, conference.ID, activeConference)
	filteredRequests := filterRequestsByConference(requests, conference.ID, activeConference)
	if len(filtered) == 0 {
		_, _ = fmt.Fprintf(out, "No booking found for active conference %q.\n", activeConference)
		return nil
	}

	renderStatusActiveConference(out, activeConference, filtered, filteredRequests)
	return nil
}

func renderStatusActiveConference(out io.Writer, activeConference string, bookings []client.BookingResponse, requests []client.RoommateRequestResponse) {
	_, _ = fmt.Fprintf(out, "Booking status for conference %q:\n", activeConference)
	for i := range bookings {
		renderBookingProjection(out, bookings[i], false, requests)
	}
}

func renderStatusAllConferences(out io.Writer, bookings []client.BookingResponse, requests []client.RoommateRequestResponse) {
	if len(bookings) == 0 {
		_, _ = fmt.Fprintln(out, "No bookings found across conferences.")
		return
	}

	sort.SliceStable(bookings, func(i, j int) bool {
		left := bookingConferenceLabel(bookings[i])
		right := bookingConferenceLabel(bookings[j])
		return left < right
	})

	_, _ = fmt.Fprintln(out, "Booking status across all conferences:")
	for i := range bookings {
		confRequests := filterRequestsByConference(requests, bookings[i].ConferenceID, bookings[i].Conference.Slug)
		renderBookingProjection(out, bookings[i], true, confRequests)
	}
}

func renderBookingProjection(out io.Writer, booking client.BookingResponse, includeConference bool, requests []client.RoommateRequestResponse) {
	if includeConference {
		_, _ = fmt.Fprintf(out, "Conference: %s\n", bookingConferenceLabel(booking))
	}

	_, _ = fmt.Fprintf(out, "  Room:       %s (%s)\n", bookingRoomNumber(booking), bookingRoomType(booking))
	_, _ = fmt.Fprintf(out, "  Price:      $%.2f/night\n", bookingPricePerNight(booking))
	_, _ = fmt.Fprintf(out, "  Dates:      %s - %s\n", bookingStartDate(booking), bookingEndDate(booking))
	_, _ = fmt.Fprintf(out, "  Privacy:    %s\n", strings.TrimSpace(booking.PrivacySetting))
	_, _ = fmt.Fprintf(out, "  Status:     %s\n", strings.TrimSpace(booking.Status))
	renderRoommates(out, booking.Roommates)
	renderPendingRequestPlaceholder(out, requests)
}

func renderRoommates(out io.Writer, roommates []client.BookingRoommateResponse) {
	if len(roommates) == 0 {
		_, _ = fmt.Fprintln(out, "  Roommates:  none")
		return
	}

	_, _ = fmt.Fprintln(out, "  Roommates:")
	for i := range roommates {
		if strings.EqualFold(strings.TrimSpace(roommates[i].PrivacySetting), "private") {
			_, _ = fmt.Fprintln(out, "    - Private attendee")
			continue
		}

		name := strings.TrimSpace(roommates[i].DisplayName)
		if name == "" {
			name = "Attendee"
		}

		_, _ = fmt.Fprintf(out, "    - %s\n", name)
	}
}

func renderPendingRequestPlaceholder(out io.Writer, requests []client.RoommateRequestResponse) {
	incoming, outgoing := pendingRequestCounts(requests)

	_, _ = fmt.Fprintln(out, "  Pending roommate requests (Epic 4 placeholder):")
	_, _ = fmt.Fprintf(out, "    Incoming: %d\n", incoming)
	_, _ = fmt.Fprintf(out, "    Outgoing: %d\n", outgoing)
	_, _ = fmt.Fprintln(out, "    Note: Full roommate request workflow arrives in Epic 4.")
}

func pendingRequestCounts(requests []client.RoommateRequestResponse) (int, int) {
	incoming := 0
	outgoing := 0

	for i := range requests {
		if !strings.EqualFold(strings.TrimSpace(requests[i].Status), "pending") {
			continue
		}

		direction := strings.ToLower(strings.TrimSpace(requests[i].Direction))
		switch direction {
		case "incoming":
			incoming++
		case "outgoing":
			outgoing++
		default:
			outgoing++
		}
	}

	return incoming, outgoing
}

func filterBookingsByConference(bookings []client.BookingResponse, conferenceID int64, conferenceSlug string) []client.BookingResponse {
	filtered := make([]client.BookingResponse, 0, len(bookings))
	normalizedSlug := strings.ToLower(strings.TrimSpace(conferenceSlug))

	for i := range bookings {
		if bookings[i].ConferenceID == conferenceID {
			filtered = append(filtered, bookings[i])
			continue
		}

		if strings.ToLower(strings.TrimSpace(bookings[i].ConferenceSlug)) == normalizedSlug {
			filtered = append(filtered, bookings[i])
			continue
		}

		if strings.ToLower(strings.TrimSpace(bookings[i].Conference.Slug)) == normalizedSlug {
			filtered = append(filtered, bookings[i])
		}
	}

	return filtered
}

func filterRequestsByConference(requests []client.RoommateRequestResponse, conferenceID int64, conferenceSlug string) []client.RoommateRequestResponse {
	filtered := make([]client.RoommateRequestResponse, 0, len(requests))
	normalizedSlug := strings.ToLower(strings.TrimSpace(conferenceSlug))

	for i := range requests {
		if requests[i].ConferenceID > 0 && conferenceID > 0 && requests[i].ConferenceID == conferenceID {
			filtered = append(filtered, requests[i])
			continue
		}

		if strings.ToLower(strings.TrimSpace(requests[i].ConferenceSlug)) == normalizedSlug {
			filtered = append(filtered, requests[i])
		}
	}

	return filtered
}

func bookingConferenceLabel(booking client.BookingResponse) string {
	name := strings.TrimSpace(booking.Conference.Name)
	slug := strings.TrimSpace(booking.Conference.Slug)

	if slug == "" {
		slug = strings.TrimSpace(booking.ConferenceSlug)
	}

	if name != "" && slug != "" {
		return fmt.Sprintf("%s (%s)", name, slug)
	}
	if name != "" {
		return name
	}
	if slug != "" {
		return slug
	}

	if booking.ConferenceID > 0 {
		return fmt.Sprintf("conference #%d", booking.ConferenceID)
	}

	return "conference"
}

func bookingRoomNumber(booking client.BookingResponse) string {
	if roomNumber := strings.TrimSpace(booking.Room.RoomNumber); roomNumber != "" {
		return roomNumber
	}

	if booking.RoomID > 0 {
		return fmt.Sprintf("room #%d", booking.RoomID)
	}

	return "unknown"
}

func bookingRoomType(booking client.BookingResponse) string {
	roomType := strings.TrimSpace(booking.Room.RoomType)
	if roomType == "" {
		return "unknown"
	}

	return roomType
}

func bookingPricePerNight(booking client.BookingResponse) float64 {
	return booking.Room.PricePerNight
}

func bookingStartDate(booking client.BookingResponse) string {
	start := strings.TrimSpace(booking.Conference.StartDate)
	if start == "" {
		return "unknown"
	}

	return start
}

func bookingEndDate(booking client.BookingResponse) string {
	end := strings.TrimSpace(booking.Conference.EndDate)
	if end == "" {
		return "unknown"
	}

	return end
}

func resolveStatusScope(ctxStore StatusContextStore, showAll bool) (string, error) {
	if showAll {
		return "", nil
	}

	conferenceSlug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}

	return conferenceSlug, nil
}
