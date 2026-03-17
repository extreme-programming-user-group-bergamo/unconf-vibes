package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// AuthClient defines the interface for device flow authentication.
type AuthClient interface {
	StartDeviceFlow(ctx context.Context) (*client.DeviceFlowResponse, error)
	ExchangeDeviceCode(ctx context.Context, deviceCode string) (*client.TokenResponse, error)
}

func newLoginCmd(authClient AuthClient, tokenStore auth.TokenStore) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub via device flow",
		Long:  "Initiates GitHub device flow authentication. Follow the on-screen instructions to sign in.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runLogin(cmd, authClient, tokenStore, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force re-login even if already authenticated")

	return cmd
}

func runLogin(cmd *cobra.Command, authClient AuthClient, tokenStore auth.TokenStore, force bool) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()

	if !force && tokenStore.HasValidToken() {
		fmt.Fprintln(out, "You are already logged in. Use --force to re-authenticate.")
		return nil
	}

	slog.Info("login: starting device flow")

	deviceResp, err := authClient.StartDeviceFlow(ctx)
	if err != nil {
		return fmt.Errorf("failed to start login: %w", err)
	}

	slog.Info("login: device code received", "user_code", deviceResp.UserCode, "expires_in", deviceResp.ExpiresIn)

	fmt.Fprintf(out, "Your device code is: %s\n", deviceResp.UserCode)
	fmt.Fprintf(out, "Please visit: %s\n", deviceResp.VerificationURI)
	fmt.Fprintln(out, "Waiting for authorization...")

	tokenResp, err := pollForToken(ctx, authClient, deviceResp)
	if err != nil {
		return err
	}

	if err := tokenStore.SaveTokens(tokenResp.AccessToken, tokenResp.RefreshToken); err != nil {
		return fmt.Errorf("failed to save authentication tokens: %w", err)
	}

	slog.Info("login: authentication successful", "user_id", tokenResp.User.ID, "display_name", tokenResp.User.DisplayName)

	fmt.Fprintf(out, "Welcome, %s! You are now logged in.\n", tokenResp.User.DisplayName)

	return nil
}

func pollForToken(ctx context.Context, apiClient AuthClient, deviceResp *client.DeviceFlowResponse) (*client.TokenResponse, error) {
	interval := time.Duration(deviceResp.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}

	deadline := time.After(time.Duration(deviceResp.ExpiresIn) * time.Second)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("login: polling cancelled by user")
			return nil, fmt.Errorf("login cancelled: %w", ctx.Err())

		case <-deadline:
			slog.Error("login: device code expired before authorization completed")
			return nil, fmt.Errorf("login failed: device code expired, please try again")

		case <-ticker.C:
			slog.Debug("login: polling for token exchange")

			tokenResp, err := apiClient.ExchangeDeviceCode(ctx, deviceResp.DeviceCode)
			if err == nil {
				return tokenResp, nil
			}

			if errors.Is(err, client.ErrAuthorizationPending) {
				continue
			}

			if errors.Is(err, client.ErrSlowDown) {
				slog.Info("login: slowing down polling interval")
				ticker.Stop()
				interval += 5 * time.Second
				ticker = time.NewTicker(interval)
				continue
			}

			if errors.Is(err, client.ErrAccessDenied) {
				slog.Error("login: authorization denied by user")
				return nil, fmt.Errorf("login failed: authorization was denied")
			}

			if errors.Is(err, client.ErrExpiredDeviceCode) {
				slog.Error("login: device code expired")
				return nil, fmt.Errorf("login failed: device code expired, please try again")
			}

			slog.Error("login: unexpected error during token exchange", "error", err)
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}
}
