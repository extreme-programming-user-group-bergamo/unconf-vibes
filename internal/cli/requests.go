package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
)

type RequestsClient interface {
	ListRoommateRequests(ctx context.Context) ([]client.RoommateRequestResponse, error)
	AcceptRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error)
	DeclineRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error)
}

func newRequestsCmd(requestsClient RequestsClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "requests",
		Short: "Manage roommate requests",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRequestsList(cmd, requestsClient)
		},
	}

	cmd.AddCommand(newRequestsAcceptCmd(requestsClient))
	cmd.AddCommand(newRequestsDeclineCmd(requestsClient))

	return cmd
}

func newRequestsAcceptCmd(requestsClient RequestsClient) *cobra.Command {
	return &cobra.Command{
		Use:   "accept <request_id>",
		Short: "Accept a roommate request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID, err := parseRequestID(args[0])
			if err != nil {
				return err
			}

			_, err = requestsClient.AcceptRoommateRequest(cmd.Context(), requestID)
			if err != nil {
				handleRequestsError(cmd, "accept request", err)
				return fmt.Errorf("failed to accept request %d: %w", requestID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Accepted roommate request %d.\n", requestID)
			return nil
		},
	}
}

func newRequestsDeclineCmd(requestsClient RequestsClient) *cobra.Command {
	return &cobra.Command{
		Use:   "decline <request_id>",
		Short: "Decline a roommate request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID, err := parseRequestID(args[0])
			if err != nil {
				return err
			}

			_, err = requestsClient.DeclineRoommateRequest(cmd.Context(), requestID)
			if err != nil {
				handleRequestsError(cmd, "decline request", err)
				return fmt.Errorf("failed to decline request %d: %w", requestID, err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Declined roommate request %d.\n", requestID)
			return nil
		},
	}
}

func runRequestsList(cmd *cobra.Command, requestsClient RequestsClient) error {
	requests, err := requestsClient.ListRoommateRequests(cmd.Context())
	if err != nil {
		handleRequestsError(cmd, "list requests", err)
		return fmt.Errorf("failed to list roommate requests: %w", err)
	}

	sort.SliceStable(requests, func(i, j int) bool {
		return requests[i].ID > requests[j].ID
	})

	incoming := make([]client.RoommateRequestResponse, 0)
	outgoing := make([]client.RoommateRequestResponse, 0)

	for i := range requests {
		if strings.EqualFold(strings.TrimSpace(requests[i].Direction), "incoming") {
			incoming = append(incoming, requests[i])
			continue
		}
		outgoing = append(outgoing, requests[i])
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Incoming requests:")
	if len(incoming) == 0 {
		_, _ = fmt.Fprintln(out, "  (none)")
	} else {
		for i := range incoming {
			_, _ = fmt.Fprintf(out, "  [%d] %s — Room %s (%s) — %s\n",
				incoming[i].ID,
				requestRequesterName(incoming[i]),
				requestRoomNumber(incoming[i]),
				requestRoomType(incoming[i]),
				requestStatusLabel(incoming[i]),
			)
		}
	}

	_, _ = fmt.Fprintln(out, "Outgoing requests:")
	if len(outgoing) == 0 {
		_, _ = fmt.Fprintln(out, "  (none)")
		return nil
	}

	for i := range outgoing {
		_, _ = fmt.Fprintf(out, "  [%d] %s — %s\n",
			outgoing[i].ID,
			requestTargetName(outgoing[i]),
			requestStatusLabel(outgoing[i]),
		)
	}

	return nil
}

func parseRequestID(raw string) (int64, error) {
	requestID, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || requestID <= 0 {
		return 0, fmt.Errorf("invalid argument: request_id must be a positive integer")
	}
	return requestID, nil
}

func handleRequestsError(cmd *cobra.Command, action string, err error) {
	errOut := cmd.ErrOrStderr()

	switch {
	case errors.Is(err, auth.ErrNotAuthenticated):
		_, _ = fmt.Fprintln(errOut, "You are not logged in. Run 'unconf login' to authenticate.")
	case errors.Is(err, client.ErrSessionExpired):
		_, _ = fmt.Fprintln(errOut, "Your session has expired. Please run 'unconf login' to re-authenticate.")
	case errors.Is(err, client.ErrUnauthorized):
		_, _ = fmt.Fprintln(errOut, "You are not authorized. Please run 'unconf login' and try again.")
	case errors.Is(err, client.ErrRequestNotFound):
		_, _ = fmt.Fprintln(errOut, "Roommate request not found.")
	case errors.Is(err, client.ErrRequestForbidden):
		_, _ = fmt.Fprintln(errOut, "You can only respond to requests addressed to you.")
	case errors.Is(err, client.ErrInvalidRequestState):
		_, _ = fmt.Fprintln(errOut, "Roommate request is already resolved.")
	default:
		_, _ = fmt.Fprintf(errOut, "Could not %s. Please try again.\n", action)
	}
}

func requestRequesterName(req client.RoommateRequestResponse) string {
	if name := strings.TrimSpace(req.RequesterName); name != "" {
		return name
	}
	return fmt.Sprintf("user #%d", req.RequesterID)
}

func requestTargetName(req client.RoommateRequestResponse) string {
	if name := strings.TrimSpace(req.TargetName); name != "" {
		return name
	}
	return fmt.Sprintf("user #%d", req.TargetID)
}

func requestRoomNumber(req client.RoommateRequestResponse) string {
	if number := strings.TrimSpace(req.RoomNumber); number != "" {
		return number
	}
	if req.RoomID > 0 {
		return fmt.Sprintf("#%d", req.RoomID)
	}
	return "unknown"
}

func requestRoomType(req client.RoommateRequestResponse) string {
	if roomType := strings.TrimSpace(req.RoomType); roomType != "" {
		return roomType
	}
	return "unknown"
}

func requestStatusLabel(req client.RoommateRequestResponse) string {
	status := strings.TrimSpace(req.Status)
	if status == "" {
		return "unknown"
	}
	return status
}
