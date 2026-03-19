package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// StatusClient defines the interface for fetching the user profile.
type StatusClient interface {
	GetMe(ctx context.Context) (*client.UserResponse, error)
}

func newStatusCmd(statusClient StatusClient) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status and user profile",
		Long:  "Displays your authenticated user profile by calling the API. Shows an error if not logged in.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd, statusClient)
		},
	}
}

func runStatus(cmd *cobra.Command, statusClient StatusClient) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("status: fetching user profile")

	user, err := statusClient.GetMe(ctx)
	if err != nil {
		if errors.Is(err, auth.ErrNotAuthenticated) {
			_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
			return fmt.Errorf("failed to get user profile: %w", err)
		}

		if errors.Is(err, client.ErrSessionExpired) {
			_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
			return fmt.Errorf("failed to get user profile: %w", err)
		}

		_, _ = fmt.Fprintln(errOut, "Could not connect to the API. Please check your connection and try again.")
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	_, _ = fmt.Fprintf(out, "Logged in as %s\n", user.DisplayName)
	_, _ = fmt.Fprintf(out, "  Email:     %s\n", user.Email)
	_, _ = fmt.Fprintf(out, "  GitHub ID: %s\n", user.GitHubID)

	return nil
}
