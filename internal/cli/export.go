package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// ExportClient defines API operations needed for organizer CSV export.
type ExportClient interface {
	ExportConferenceBookingsCSV(ctx context.Context, slug string, includeCancelled bool) ([]byte, error)
}

// ExportContextStore defines the interface for reading active conference context.
type ExportContextStore interface {
	GetActiveConference() (string, error)
}

func newExportCmd(exportClient ExportClient, ctxStore ExportContextStore) *cobra.Command {
	var conferenceSlug string
	var outputPath string
	var includeCancelled bool

	cmd := &cobra.Command{
		Use:   "export [conference-slug]",
		Short: "Export conference bookings as CSV (organizer only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveExportSlug(cmd, ctxStore, conferenceSlug, args)
			if err != nil || slug == "" {
				return err
			}

			csvData, err := exportClient.ExportConferenceBookingsCSV(cmd.Context(), slug, includeCancelled)
			if err != nil {
				return handleExportFetchError(cmd, slug, err)
			}

			if strings.TrimSpace(outputPath) == "" {
				_, _ = cmd.OutOrStdout().Write(csvData)
				return nil
			}

			if err := os.WriteFile(strings.TrimSpace(outputPath), csvData, 0o600); err != nil {
				return fmt.Errorf("failed to write export file: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Exported conference bookings CSV to %s\n", strings.TrimSpace(outputPath))
			return nil
		},
	}

	cmd.Flags().StringVar(&conferenceSlug, "conference", "", "Conference slug (defaults to active context)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write CSV output to file path")
	cmd.Flags().BoolVar(&includeCancelled, "include-cancelled", false, "Include cancelled bookings in export")

	return cmd
}

func resolveExportSlug(cmd *cobra.Command, ctxStore ExportContextStore, explicitSlug string, args []string) (string, error) {
	if slug := strings.TrimSpace(explicitSlug); slug != "" {
		return slug, nil
	}
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
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Use --conference <slug>, provide [conference-slug], or set context with 'unconf checkout <slug>'")
		return "", nil
	}

	return strings.TrimSpace(slug), nil
}

func handleExportFetchError(cmd *cobra.Command, conferenceSlug string, err error) error {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return fmt.Errorf("failed to export bookings for %q: %w", conferenceSlug, auth.ErrNotAuthenticated)
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
		return fmt.Errorf("failed to export bookings for %q: %w", conferenceSlug, client.ErrSessionExpired)
	case errors.Is(err, client.ErrOrganizerForbidden):
		_, _ = fmt.Fprintln(errOut, "Organizer permissions required for CSV export.")
		return fmt.Errorf("failed to export bookings for %q: %w", conferenceSlug, client.ErrOrganizerForbidden)
	case errors.Is(err, client.ErrConferenceNotFound):
		_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", conferenceSlug)
		return fmt.Errorf("failed to export bookings for %q: %w", conferenceSlug, client.ErrConferenceNotFound)
	default:
		_, _ = fmt.Fprintln(errOut, "Could not export conference bookings from the API. Please try again.")
		return fmt.Errorf("failed to export bookings for %q: %w", conferenceSlug, err)
	}
}
