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
	GetMe(ctx context.Context, accessToken string) (*client.UserResponse, error)
}

func newStatusCmd(statusClient StatusClient, tokenStore auth.TokenStore) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status and user profile",
		Long:  "Displays your authenticated user profile by calling the API. Shows an error if not logged in.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd, statusClient, tokenStore)
		},
	}
}

func runStatus(cmd *cobra.Command, statusClient StatusClient, tokenStore auth.TokenStore) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	if err := auth.RequireAuth(tokenStore); err != nil {
		fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return err
	}

	accessToken, err := tokenStore.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to retrieve access token: %w", err)
	}

	slog.Info("status: fetching user profile")

	user, err := statusClient.GetMe(ctx, accessToken)
	if err != nil {
		if errors.Is(err, client.ErrUnauthorized) {
			slog.Warn("status: token rejected by server, clearing local tokens")

			if clearErr := tokenStore.ClearTokens(); clearErr != nil {
				slog.Error("status: failed to clear tokens", "error", clearErr)
			}

			fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")

			return fmt.Errorf("failed to get user profile: %w", err)
		}

		fmt.Fprintln(errOut, "Could not connect to the API. Please check your connection and try again.")

		return fmt.Errorf("failed to get user profile: %w", err)
	}

	fmt.Fprintf(out, "Logged in as %s\n", user.DisplayName)
	fmt.Fprintf(out, "  Email:     %s\n", user.Email)
	fmt.Fprintf(out, "  GitHub ID: %s\n", user.GitHubID)

	return nil
}
