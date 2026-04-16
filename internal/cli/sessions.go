package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type SessionsClient interface {
	ListSessions(ctx context.Context) ([]client.SessionResponse, error)
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeOtherSessions(ctx context.Context) (int64, error)
}

func newSessionsCmd(sessionsClient SessionsClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage active authenticated sessions",
	}

	cmd.AddCommand(newSessionsListCmd(sessionsClient))
	cmd.AddCommand(newSessionsRevokeCmd(sessionsClient))
	cmd.AddCommand(newSessionsRevokeOthersCmd(sessionsClient))

	return cmd
}

func newSessionsListCmd(sessionsClient SessionsClient) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sessions, err := sessionsClient.ListSessions(cmd.Context())
			if err != nil {
				return handleSessionsAuthError(cmd.ErrOrStderr(), "failed to list sessions", err)
			}

			out := cmd.OutOrStdout()
			if len(sessions) == 0 {
				_, _ = fmt.Fprintln(out, "No active sessions found.")
				return nil
			}

			_, _ = fmt.Fprintln(out, "Active Sessions")
			_, _ = fmt.Fprintln(out, "===============")
			for _, session := range sessions {
				currentFlag := ""
				if session.Current {
					currentFlag = " (current)"
				}
				_, _ = fmt.Fprintf(
					out,
					"- ID: %d%s\n  Created: %s\n  Last Seen: %s\n  Client: %s\n",
					session.ID,
					currentFlag,
					session.CreatedAt.Format(time.RFC3339),
					session.LastSeenAt.Format(time.RFC3339),
					emptyToUnknown(session.ClientMetadata),
				)
			}

			return nil
		},
	}
}

func newSessionsRevokeCmd(sessionsClient SessionsClient) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke [session-id]",
		Short: "Revoke a specific session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || sessionID <= 0 {
				return fmt.Errorf("invalid session id %q: must be a positive integer", args[0])
			}

			if err := sessionsClient.RevokeSession(cmd.Context(), sessionID); err != nil {
				if errors.Is(err, client.ErrCurrentSessionRevoke) {
					_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Cannot revoke your current session with this command. Use 'unconf logout' instead.")
				}
				return handleSessionsAuthError(cmd.ErrOrStderr(), "failed to revoke session", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session %d revoked.\n", sessionID)
			return nil
		},
	}
}

func newSessionsRevokeOthersCmd(sessionsClient SessionsClient) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "revoke-others",
		Short: "Revoke all other sessions except the current one",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !force {
				confirmed, err := confirmSessionsRevokeOthers(cmd.InOrStdin(), cmd.OutOrStdout())
				if err != nil {
					return err
				}
				if !confirmed {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			revokedCount, err := sessionsClient.RevokeOtherSessions(cmd.Context())
			if err != nil {
				return handleSessionsAuthError(cmd.ErrOrStderr(), "failed to revoke other sessions", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Revoked %d other session(s).\n", revokedCount)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "yes", false, "Skip confirmation prompt")
	return cmd
}

func confirmSessionsRevokeOthers(in io.Reader, out io.Writer) (bool, error) {
	_, _ = fmt.Fprint(out, "Revoke all other sessions? [y/N]: ")
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, fmt.Errorf("failed to read confirmation: %w", err)
		}
		return false, nil
	}

	choice := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return choice == "y" || choice == "yes", nil
}

func handleSessionsAuthError(errOut io.Writer, msg string, err error) error {
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

func emptyToUnknown(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}
	return trimmed
}
