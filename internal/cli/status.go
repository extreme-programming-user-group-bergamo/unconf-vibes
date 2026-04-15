package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
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

	scope := activeConference
	if showAll {
		scope = "all"
	}
	slog.Info("status: fetching booking status", "scope", scope)

	bookings, err := statusClient.ListBookings(ctx)
	if err != nil {
		return handleStatusFetchError(cmd, err, "bookings")
	}
	bookings = filterActiveBookingResponses(bookings)

	requests, err := statusClient.ListRoommateRequests(ctx)
	if err != nil {
		return handleStatusFetchError(cmd, err, "roommate requests")
	}

	if showAll {
		renderStatusAllConferences(out, bookings, requests)
		slog.Info("status: rendered booking status", "scope", "all", "bookings", len(bookings), "requests", len(requests))
		return nil
	}

	conference, confErr := statusClient.GetConference(ctx, activeConference)
	if confErr != nil {
		if errors.Is(confErr, client.ErrConferenceNotFound) {
			slog.Error("status: active conference context not found", "conference_slug", activeConference, "error", confErr)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Active conference %q was not found. Run 'unconf checkout <slug>' to switch context.\n", activeConference)
			return fmt.Errorf("failed to resolve conference context for %q: %w", activeConference, confErr)
		}

		slog.Error("status: failed to resolve active conference context", "conference_slug", activeConference, "error", confErr)
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Could not resolve active conference context. Please try again.")
		return fmt.Errorf("failed to resolve conference context for %q: %w", activeConference, confErr)
	}

	filtered := filterBookingsByConference(bookings, conference.ID, activeConference)
	filteredRequests := filterRequestsByConference(requests, conference.ID, activeConference)
	if len(filtered) == 0 {
		_, _ = fmt.Fprintf(out, "No booking found for active conference %q. Run 'unconf book <room_number>' to create one, or use 'unconf status --all'.\n", activeConference)
		slog.Info("status: no booking found for active conference", "conference_slug", activeConference)
		return nil
	}

	renderStatusActiveConference(out, activeConference, filtered, filteredRequests)
	slog.Info("status: rendered booking status", "scope", activeConference, "bookings", len(filtered), "requests", len(filteredRequests))
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

	_, _ = fmt.Fprintf(out, "  Room:       %s\n", bookingRoomSummary(booking))
	if booking.Room.Capacity > 0 {
		_, _ = fmt.Fprintf(out, "  Type:       %s\n", bookingRoomType(booking))
	}
	_, _ = fmt.Fprintf(out, "  Price:      $%.2f/night\n", bookingPricePerNight(booking))
	_, _ = fmt.Fprintf(out, "  Dates:      %s - %s\n", bookingStartDate(booking), bookingEndDate(booking))
	_, _ = fmt.Fprintf(out, "  Privacy:    %s\n", strings.TrimSpace(booking.PrivacySetting))
	_, _ = fmt.Fprintf(out, "  Status:     %s\n", strings.TrimSpace(booking.Status))
	renderRoommates(out, booking.Roommates)
	renderRequestSummary(out, requests)
}

func bookingRoomSummary(booking client.BookingResponse) string {
	roomNumber := bookingRoomNumber(booking)
	if booking.Room.Capacity <= 0 {
		return fmt.Sprintf("%s (%s)", roomNumber, bookingRoomType(booking))
	}

	spotsTaken := booking.Room.SpotsTaken
	if spotsTaken < 0 {
		spotsTaken = 0
	}

	summary := fmt.Sprintf("Room %s (%d/%d spots)", roomNumber, spotsTaken, booking.Room.Capacity)
	openSpots := booking.Room.Capacity - spotsTaken
	if openSpots == 1 {
		return summary + " - open spot available"
	}
	if openSpots > 1 {
		return fmt.Sprintf("%s - %d open spots available", summary, openSpots)
	}

	return summary
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

func renderRequestSummary(out io.Writer, requests []client.RoommateRequestResponse) {
	incoming, outgoing := pendingRequestCounts(requests)
	updates := resolvedRequestUpdates(requests)

	_, _ = fmt.Fprintln(out, "  Roommate requests:")
	_, _ = fmt.Fprintf(out, "    Pending incoming: %d\n", incoming)
	_, _ = fmt.Fprintf(out, "    Pending outgoing: %d\n", outgoing)
	if len(updates) == 0 {
		return
	}

	_, _ = fmt.Fprintln(out, "    Recent updates:")
	for i := range updates {
		_, _ = fmt.Fprintf(out, "      - %s\n", updates[i])
	}
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

func resolvedRequestUpdates(requests []client.RoommateRequestResponse) []string {
	updates := make([]string, 0)

	for i := range requests {
		status := strings.ToLower(strings.TrimSpace(requests[i].Status))
		if status != "accepted" && status != "declined" && status != "cancelled" {
			continue
		}

		direction := strings.ToLower(strings.TrimSpace(requests[i].Direction))
		switch direction {
		case "incoming":
			updates = append(updates, fmt.Sprintf("%s %s's request for room %s",
				titleWord(status),
				requestRequesterName(requests[i]),
				requestRoomNumber(requests[i]),
			))
		case "outgoing":
			updates = append(updates, fmt.Sprintf("Your request to %s was %s",
				requestTargetName(requests[i]),
				status,
			))
		default:
			updates = append(updates, fmt.Sprintf("Request %d is %s", requests[i].ID, status))
		}
	}

	sort.Strings(updates)
	return updates
}

func titleWord(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return trimmed
	}

	return strings.ToUpper(trimmed[:1]) + trimmed[1:]
}

func filterBookingsByConference(bookings []client.BookingResponse, conferenceID int64, conferenceSlug string) []client.BookingResponse {
	filtered := make([]client.BookingResponse, 0, len(bookings))
	normalizedSlug := strings.ToLower(strings.TrimSpace(conferenceSlug))

	for i := range bookings {
		if strings.EqualFold(strings.TrimSpace(bookings[i].Status), "cancelled") {
			continue
		}

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

func filterActiveBookingResponses(bookings []client.BookingResponse) []client.BookingResponse {
	filtered := make([]client.BookingResponse, 0, len(bookings))

	for i := range bookings {
		if strings.EqualFold(strings.TrimSpace(bookings[i].Status), "cancelled") {
			continue
		}

		filtered = append(filtered, bookings[i])
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

func handleStatusFetchError(cmd *cobra.Command, err error, resource string) error {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return fmt.Errorf("failed to fetch %s: %w", resource, auth.ErrNotAuthenticated)
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
		return fmt.Errorf("failed to fetch %s: %w", resource, client.ErrSessionExpired)
	case errors.Is(err, client.ErrUnauthorized):
		_, _ = fmt.Fprintln(errOut, "You are not authorized. Please run 'unconf login' and try again.")
		return fmt.Errorf("failed to fetch %s: %w", resource, client.ErrUnauthorized)
	default:
		slog.Error("status: status fetch failed", "resource", resource, "error", err)
		_, _ = fmt.Fprintln(errOut, "Could not fetch status details from the API. Please try again.")
		return fmt.Errorf("failed to fetch %s: %w", resource, err)
	}
}
