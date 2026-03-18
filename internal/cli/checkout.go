package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// CheckoutClient defines the interface for validating a conference slug.
type CheckoutClient interface {
	GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error)
}

// CheckoutContextStore defines the interface for reading/writing active conference context.
type CheckoutContextStore interface {
	GetActiveConference() (string, error)
	SetActiveConference(slug string) error
}

func newCheckoutCmd(checkoutClient CheckoutClient, ctxStore CheckoutContextStore) *cobra.Command {
	return &cobra.Command{
		Use:   "checkout [slug]",
		Short: "Set active conference context",
		Long:  "Sets the active conference context so that subsequent commands operate on that conference. Run without arguments to see the current context.",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runCheckoutShow(cmd, ctxStore)
			}
			return runCheckoutSet(cmd, checkoutClient, ctxStore, args[0])
		},
	}
}

func runCheckoutShow(cmd *cobra.Command, ctxStore CheckoutContextStore) error {
	out := cmd.OutOrStdout()

	slug, err := ctxStore.GetActiveConference()
	if err != nil {
		return fmt.Errorf("failed to read current context: %w", err)
	}

	if slug == "" {
		_, _ = fmt.Fprintln(out, "No conference selected. Use 'unconf checkout <slug>' to set one.")
		return nil
	}

	_, _ = fmt.Fprintf(out, "Current conference: %s\n", slug)
	return nil
}

func runCheckoutSet(cmd *cobra.Command, checkoutClient CheckoutClient, ctxStore CheckoutContextStore, slug string) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("checkout: validating conference", "slug", slug)

	_, err := checkoutClient.GetConference(ctx, slug)
	if err != nil {
		if errors.Is(err, client.ErrConferenceNotFound) {
			_, _ = fmt.Fprintf(errOut, "Conference %q not found. Use 'unconf list' to see available conferences.\n", slug)
			return fmt.Errorf("failed to checkout: %w", err)
		}
		return fmt.Errorf("failed to validate conference: %w", err)
	}

	if err := ctxStore.SetActiveConference(slug); err != nil {
		return fmt.Errorf("failed to save conference context: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Switched to %s\n", slug)
	return nil
}
