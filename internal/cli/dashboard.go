package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/tui/common"
	tuidashboard "github.com/katurdays/unconf/internal/tui/dashboard"
	"github.com/spf13/cobra"
)

// DashboardClient defines API operations needed for organizer dashboard.
type DashboardClient interface {
	GetOrganizerDashboard(ctx context.Context, slug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error)
}

// DashboardContextStore defines the interface for reading active conference context.
type DashboardContextStore interface {
	GetActiveConference() (string, error)
}

var launchDashboardProgram = func(model tea.Model) (tea.Model, error) {
	return tea.NewProgram(model).Run()
}

func newDashboardCmd(dashboardClient DashboardClient, ctxStore DashboardContextStore) *cobra.Command {
	return newDashboardCmdWithChecker(dashboardClient, ctxStore, defaultTerminalCapabilityChecker{})
}

func newDashboardCmdWithChecker(
	dashboardClient DashboardClient,
	ctxStore DashboardContextStore,
	checker terminalCapabilityChecker,
) *cobra.Command {
	var roomType string
	var bookingStatus string
	var hasSpecialRequests bool
	var search string

	cmd := &cobra.Command{
		Use:   "dashboard [conference-slug]",
		Short: "Open organizer dashboard",
		Long:  "Launches the organizer dashboard TUI. If no slug is provided, uses the active conference context.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveDashboardSlug(cmd, ctxStore, args)
			if err != nil {
				return err
			}
			if slug == "" {
				return nil
			}

			query := client.DashboardQuery{
				RoomType:           strings.TrimSpace(roomType),
				BookingStatus:      strings.TrimSpace(bookingStatus),
				HasSpecialRequests: hasSpecialRequests,
				Search:             strings.TrimSpace(search),
			}
			return runDashboard(cmd, dashboardClient, checker, slug, query)
		},
	}

	cmd.Flags().StringVar(&roomType, "room-type", "", "Optional initial room-type filter (single|double|triple)")
	cmd.Flags().StringVar(&bookingStatus, "booking-status", "", "Optional initial booking status filter (requested|confirmed|cancelled)")
	cmd.Flags().BoolVar(&hasSpecialRequests, "has-special-requests", false, "Show only attendees with dietary/accessibility notes")
	cmd.Flags().StringVar(&search, "search", "", "Optional attendee name search term")

	return cmd
}

func resolveDashboardSlug(cmd *cobra.Command, ctxStore DashboardContextStore, args []string) (string, error) {
	if len(args) == 1 {
		if slug := strings.TrimSpace(args[0]); slug != "" {
			return slug, nil
		}
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}
	if strings.TrimSpace(slug) == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Usage: unconf dashboard <slug> or set context with 'unconf checkout <slug>'")
		return "", nil
	}

	return strings.TrimSpace(slug), nil
}

func runDashboard(
	cmd *cobra.Command,
	dashboardClient DashboardClient,
	checker terminalCapabilityChecker,
	slug string,
	query client.DashboardQuery,
) error {
	if !checker.SupportsInteractiveUI() {
		slog.Warn("dashboard: interactive terminal unsupported, falling back to plain output", "slug", slug)
		return runDashboardFallback(cmd, dashboardClient, slug, query)
	}

	model := tuidashboard.NewModel(
		cmd.Context(),
		slug,
		func(ctx context.Context, conferenceSlug string, q client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
			return dashboardClient.GetOrganizerDashboard(ctx, conferenceSlug, q)
		},
		query,
		common.NewStyles(),
	)
	if _, err := launchDashboardProgram(model); err != nil {
		return fmt.Errorf("failed to launch organizer dashboard: %w", err)
	}

	return nil
}

func runDashboardFallback(cmd *cobra.Command, dashboardClient DashboardClient, slug string, query client.DashboardQuery) error {
	result, err := dashboardClient.GetOrganizerDashboard(cmd.Context(), slug, query)
	if err != nil {
		return handleDashboardFetchError(cmd, slug, err)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Interactive organizer dashboard unavailable in this terminal.")
	_, _ = fmt.Fprintf(out, "Showing plain dashboard summary for conference %q.\n\n", slug)
	_, _ = fmt.Fprintf(out, "Total registrations: %d\n", result.TotalRegistrations)
	_, _ = fmt.Fprintf(out, "Capacity usage: %.1f%% (%d capacity)\n\n", result.CapacityUsagePct, result.Capacity)

	if len(result.RoomFillRates) > 0 {
		_, _ = fmt.Fprintln(out, "Room fill rates:")
		ratesWriter := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(ratesWriter, "ROOM TYPE\tREGISTRATIONS\tCAPACITY\tFILL RATE")
		for i := range result.RoomFillRates {
			rate := result.RoomFillRates[i]
			_, _ = fmt.Fprintf(ratesWriter, "%s\t%d\t%d\t%.1f%%\n", rate.RoomType, rate.Registrations, rate.Capacity, rate.FillRatePct)
		}
		_ = ratesWriter.Flush()
		_, _ = fmt.Fprintln(out)
	}

	if len(result.Attendees) == 0 {
		_, _ = fmt.Fprintln(out, "No attendees match current filters.")
		_, _ = fmt.Fprintln(out, "Tip: use an interactive terminal (TTY with TERM set) to launch full TUI mode.")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tEMAIL\tROOM\tSTATUS\tDIETARY/ACCESSIBILITY\tPRIVACY")
	for i := range result.Attendees {
		row := result.Attendees[i]
		room := strings.TrimSpace(row.RoomNumber)
		if room == "" {
			room = "not booked yet"
		}
		if strings.TrimSpace(row.RoomType) != "" {
			room = fmt.Sprintf("%s (%s)", room, row.RoomType)
		}
		note := strings.TrimSpace(row.DietaryAccessibilityNote)
		if note == "" {
			note = "-"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", row.Name, row.Email, room, row.BookingStatus, note, row.PrivacySetting)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to render dashboard fallback table: %w", err)
	}

	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "Tip: use an interactive terminal (TTY with TERM set) to launch full TUI mode.")
	return nil
}

func handleDashboardFetchError(cmd *cobra.Command, conferenceSlug string, err error) error {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return fmt.Errorf("failed to load dashboard for %q: %w", conferenceSlug, auth.ErrNotAuthenticated)
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
		return fmt.Errorf("failed to load dashboard for %q: %w", conferenceSlug, client.ErrSessionExpired)
	case errors.Is(err, client.ErrOrganizerForbidden):
		_, _ = fmt.Fprintln(errOut, "Organizer permissions required for dashboard access.")
		return fmt.Errorf("failed to load dashboard for %q: %w", conferenceSlug, client.ErrOrganizerForbidden)
	case errors.Is(err, client.ErrConferenceNotFound):
		_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", conferenceSlug)
		return fmt.Errorf("failed to load dashboard for %q: %w", conferenceSlug, client.ErrConferenceNotFound)
	default:
		_, _ = fmt.Fprintln(errOut, "Could not fetch organizer dashboard from the API. Please try again.")
		return fmt.Errorf("failed to load dashboard for %q: %w", conferenceSlug, err)
	}
}
