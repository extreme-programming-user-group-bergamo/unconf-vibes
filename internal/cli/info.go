package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// InfoClient defines the interface for fetching a single conference.
type InfoClient interface {
	GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error)
}

// InfoContextStore defines the interface for reading active conference context.
type InfoContextStore interface {
	GetActiveConference() (string, error)
}

func newInfoCmd(infoClient InfoClient, ctxStore InfoContextStore) *cobra.Command {
	return &cobra.Command{
		Use:   "info [conference-slug]",
		Short: "Show detailed conference information",
		Long:  "Displays detailed information about a specific conference including dates, location, capacity, and room availability.\nIf no slug is provided, uses the active conference context set via 'unconf checkout'.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug, err := resolveInfoSlug(cmd, ctxStore, args)
			if err != nil {
				return err
			}
			if slug == "" {
				return nil
			}
			return runInfo(cmd, infoClient, slug)
		},
	}
}

func resolveInfoSlug(cmd *cobra.Command, ctxStore InfoContextStore, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return "", fmt.Errorf("failed to read conference context: %w", err)
	}

	if slug == "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No conference specified. Usage: unconf info <slug> or set context with 'unconf checkout <slug>'")
		return "", nil
	}

	return slug, nil
}

func runInfo(cmd *cobra.Command, infoClient InfoClient, slug string) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("info: fetching conference", "slug", slug)

	conf, err := infoClient.GetConference(ctx, slug)
	if err != nil {
		if errors.Is(err, client.ErrConferenceNotFound) {
			_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", slug)
			return fmt.Errorf("failed to get conference info: %w", err)
		}
		return fmt.Errorf("failed to get conference info: %w", err)
	}

	dates := formatDateRange(conf.StartDate, conf.EndDate)

	// Conference name header
	_, _ = fmt.Fprintln(out, conf.Name)
	_, _ = fmt.Fprintln(out, strings.Repeat("═", 42))
	_, _ = fmt.Fprintln(out)

	// Description
	_, _ = fmt.Fprintln(out, conf.Description)
	_, _ = fmt.Fprintln(out)

	// Details
	_, _ = fmt.Fprintf(out, "  Location:    %s\n", conf.Location)
	_, _ = fmt.Fprintf(out, "  Dates:       %s\n", dates)
	_, _ = fmt.Fprintf(out, "  Status:      %s\n", conf.Status)
	_, _ = fmt.Fprintf(out, "  Capacity:    %d\n", conf.Capacity)
	_, _ = fmt.Fprintf(out, "  Attendees:   %d\n", conf.AttendeeCount)
	_, _ = fmt.Fprintln(out)

	// Rooms section (placeholder until Epic 3)
	_, _ = fmt.Fprintln(out, "Rooms")
	_, _ = fmt.Fprintln(out, strings.Repeat("─", 42))
	_, _ = fmt.Fprintln(out, "No room information available yet.")

	return nil
}
