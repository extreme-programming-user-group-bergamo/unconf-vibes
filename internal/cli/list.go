package cli

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// ListClient defines the interface for fetching conferences.
type ListClient interface {
	ListConferences(ctx context.Context) ([]client.ConferenceResponse, error)
}

func newListCmd(listClient ListClient) *cobra.Command {
	var showAll bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available conferences",
		Long:  "Displays conferences in a formatted table. By default only upcoming and active conferences are shown. Use --all to include past conferences.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runList(cmd, listClient, showAll)
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show past conferences as well")

	return cmd
}

func runList(cmd *cobra.Command, listClient ListClient, showAll bool) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()

	slog.Info("list: fetching conferences")

	conferences, err := listClient.ListConferences(ctx)
	if err != nil {
		return fmt.Errorf("failed to list conferences: %w", err)
	}

	// Filter out past conferences unless --all is set.
	var filtered []client.ConferenceResponse
	for _, c := range conferences {
		if showAll || c.Status != "past" {
			filtered = append(filtered, c)
		}
	}

	// Sort by start_date ascending.
	slices.SortFunc(filtered, func(a, b client.ConferenceResponse) int {
		ta, _ := time.Parse("2006-01-02", a.StartDate)
		tb, _ := time.Parse("2006-01-02", b.StartDate)
		return ta.Compare(tb)
	})

	if len(filtered) == 0 {
		if showAll {
			_, _ = fmt.Fprintln(out, "No conferences available.")
		} else {
			_, _ = fmt.Fprintln(out, "No upcoming conferences. Use --all to show past conferences.")
		}
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tDATES\tLOCATION\tATTENDEES\tSTATUS")

	for _, c := range filtered {
		dates := formatDateRange(c.StartDate, c.EndDate)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", c.Name, dates, c.Location, c.AttendeeCount, c.Status)
	}

	return w.Flush()
}

func formatDateRange(startStr, endStr string) string {
	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return startStr + " - " + endStr
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return startStr + " - " + endStr
	}

	return start.Format("Jan 02") + " - " + end.Format("Jan 02, 2006")
}
