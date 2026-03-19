package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

// ConfigClient defines the interface for profile operations.
type ConfigClient interface {
	GetMe(ctx context.Context) (*client.UserResponse, error)
	UpdateMe(ctx context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error)
}

func newConfigCmd(configClient ConfigClient) *cobra.Command {
	var (
		showFlag    bool
		displayName string
		privacy     string
	)

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure your profile settings",
		Long:  "View and update your display name and privacy preferences. Run without flags for interactive mode.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if showFlag {
				return runConfigShow(cmd, configClient)
			}
			if cmd.Flags().Changed("display-name") || cmd.Flags().Changed("privacy") {
				return runConfigUpdate(cmd, configClient, displayName, privacy)
			}
			return runConfigInteractive(cmd, configClient)
		},
	}

	cmd.Flags().BoolVar(&showFlag, "show", false, "Display current profile settings")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Set display name")
	cmd.Flags().StringVar(&privacy, "privacy", "", "Set privacy setting (public or private)")

	return cmd
}

func runConfigShow(cmd *cobra.Command, configClient ConfigClient) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("config: fetching profile")

	user, err := configClient.GetMe(ctx)
	if err != nil {
		return handleConfigAuthError(errOut, "failed to get profile", err)
	}

	_, _ = fmt.Fprintln(out, "Profile Settings")
	_, _ = fmt.Fprintln(out, "══════════════════════════════════════════")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "  Display Name:  %s\n", user.DisplayName)
	_, _ = fmt.Fprintf(out, "  Email:         %s\n", user.Email)
	_, _ = fmt.Fprintf(out, "  Privacy:       %s\n", user.PrivacySetting)

	return nil
}

func runConfigUpdate(cmd *cobra.Command, configClient ConfigClient, displayName string, privacy string) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("config: updating profile")

	input := client.UpdateProfileRequest{}
	if cmd.Flags().Changed("display-name") {
		input.DisplayName = &displayName
	}
	if cmd.Flags().Changed("privacy") {
		if privacy != "public" && privacy != "private" {
			_, _ = fmt.Fprintln(errOut, "Invalid privacy setting. Use 'public' or 'private'.")
			return fmt.Errorf("failed to update profile: invalid privacy setting %q", privacy)
		}
		input.PrivacySetting = &privacy
	}

	_, err := configClient.UpdateMe(ctx, input)
	if err != nil {
		return handleConfigAuthError(errOut, "failed to update profile", err)
	}

	_, _ = fmt.Fprintln(out, "Profile updated successfully.")
	return nil
}

func runConfigInteractive(cmd *cobra.Command, configClient ConfigClient) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	slog.Info("config: starting interactive mode")

	user, err := configClient.GetMe(ctx)
	if err != nil {
		return handleConfigAuthError(errOut, "failed to get profile", err)
	}

	_, _ = fmt.Fprintln(out, "Current profile:")
	_, _ = fmt.Fprintf(out, "  Display Name: %s\n", user.DisplayName)
	_, _ = fmt.Fprintf(out, "  Email:        %s\n", user.Email)
	_, _ = fmt.Fprintf(out, "  Privacy:      %s\n", user.PrivacySetting)
	_, _ = fmt.Fprintln(out)

	scanner := bufio.NewScanner(cmd.InOrStdin())

	_, _ = fmt.Fprintf(out, "Enter new display name (press Enter to keep %q): ", user.DisplayName)
	var newDisplayName string
	if scanner.Scan() {
		newDisplayName = strings.TrimSpace(scanner.Text())
	}

	_, _ = fmt.Fprintf(out, "Select privacy setting [public/private] (press Enter to keep %q): ", user.PrivacySetting)
	var newPrivacy string
	if scanner.Scan() {
		newPrivacy = strings.TrimSpace(scanner.Text())
	}

	if newPrivacy != "" && newPrivacy != "public" && newPrivacy != "private" {
		_, _ = fmt.Fprintln(errOut, "Invalid privacy setting. Use 'public' or 'private'.")
		return fmt.Errorf("failed to update profile: invalid privacy setting %q", newPrivacy)
	}

	input := client.UpdateProfileRequest{}
	changed := false

	if newDisplayName != "" && newDisplayName != user.DisplayName {
		input.DisplayName = &newDisplayName
		changed = true
	}
	if newPrivacy != "" && newPrivacy != user.PrivacySetting {
		input.PrivacySetting = &newPrivacy
		changed = true
	}

	if !changed {
		_, _ = fmt.Fprintln(out, "No changes made.")
		return nil
	}

	_, err = configClient.UpdateMe(ctx, input)
	if err != nil {
		return handleConfigAuthError(errOut, "failed to update profile", err)
	}

	_, _ = fmt.Fprintln(out, "Profile updated successfully.")
	return nil
}

func handleConfigAuthError(errOut io.Writer, msg string, err error) error {
	if errors.Is(err, auth.ErrNotAuthenticated) {
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
		return fmt.Errorf("%s: %w", msg, err)
	}

	if errors.Is(err, client.ErrSessionExpired) {
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
		return fmt.Errorf("%s: %w", msg, err)
	}

	return fmt.Errorf("%s: %w", msg, err)
}
