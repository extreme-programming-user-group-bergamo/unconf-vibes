package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// StatusClient defines the interface used by the status command.
type StatusClient interface{}

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

func runStatus(cmd *cobra.Command, _ StatusClient, ctxStore StatusContextStore, showAll bool) error {
	out := cmd.OutOrStdout()

	activeConference, err := resolveStatusScope(ctxStore, showAll)
	if err != nil {
		return err
	}

	if !showAll && activeConference == "" {
		_, _ = fmt.Fprintln(out, "No active conference context. Run 'unconf checkout <slug>' or use 'unconf status --all'.")
		return nil
	}

	if showAll {
		_, _ = fmt.Fprintln(out, "Checking booking status across all conferences...")
		return nil
	}

	_, _ = fmt.Fprintf(out, "Checking booking status for active conference %q...\n", activeConference)
	return nil
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
