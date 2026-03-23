package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/spf13/cobra"
)

// RevokeClient defines the interface for token revocation.
type RevokeClient interface {
	RevokeToken(ctx context.Context, accessToken string) error
}

func newLogoutCmd(revokeClient RevokeClient, tokenStore auth.TokenStore) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out and remove stored credentials",
		Long:  "Revokes your authentication token and removes stored credentials from the OS keychain.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLogout(cmd, revokeClient, tokenStore)
		},
	}
}

func runLogout(cmd *cobra.Command, revokeClient RevokeClient, tokenStore auth.TokenStore) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()

	accessToken, err := tokenStore.GetAccessToken()
	if err != nil {
		if errors.Is(err, auth.ErrNotAuthenticated) {
			_, _ = fmt.Fprintln(out, "You are not currently logged in.")
			return nil
		}

		return fmt.Errorf("failed to retrieve access token: %w", err)
	}

	if accessToken == "" {
		_, _ = fmt.Fprintln(out, "You are not currently logged in.")
		return nil
	}

	slog.Info("logout: revoking token on server")

	if revokeErr := revokeClient.RevokeToken(ctx, accessToken); revokeErr != nil {
		slog.Warn("logout: failed to revoke token on server (continuing with local cleanup)", "error", revokeErr)
	}

	if err := tokenStore.ClearTokens(); err != nil {
		return fmt.Errorf("failed to clear stored credentials: %w", err)
	}

	slog.Info("logout: credentials cleared successfully")

	_, _ = fmt.Fprintln(out, "You have been logged out.")

	return nil
}
