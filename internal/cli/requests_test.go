package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRequestsClient struct {
	listFn    func(context.Context) ([]client.RoommateRequestResponse, error)
	acceptFn  func(context.Context, int64) (*client.RoommateRequestResponse, error)
	declineFn func(context.Context, int64) (*client.RoommateRequestResponse, error)
}

func (m *mockRequestsClient) ListRoommateRequests(ctx context.Context) ([]client.RoommateRequestResponse, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return []client.RoommateRequestResponse{}, nil
}

func (m *mockRequestsClient) AcceptRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
	if m.acceptFn != nil {
		return m.acceptFn(ctx, requestID)
	}
	return &client.RoommateRequestResponse{ID: requestID, Status: "accepted"}, nil
}

func (m *mockRequestsClient) DeclineRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
	if m.declineFn != nil {
		return m.declineFn(ctx, requestID)
	}
	return &client.RoommateRequestResponse{ID: requestID, Status: "declined"}, nil
}

func executeRequestsCmd(t *testing.T, cmd *cobra.Command, args []string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestRequestsCmd_ListsIncomingAndOutgoing(t *testing.T) {
	cmd := newRequestsCmd(&mockRequestsClient{
		listFn: func(_ context.Context) ([]client.RoommateRequestResponse, error) {
			return []client.RoommateRequestResponse{
				{ID: 3, Direction: "incoming", RequesterName: "Alice", RoomNumber: "401", RoomType: "double", Status: "pending"},
				{ID: 4, Direction: "outgoing", TargetName: "Bob", Status: "accepted"},
			}, nil
		},
	})

	stdout, stderr, err := executeRequestsCmd(t, cmd, []string{})
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Incoming requests:")
	assert.Contains(t, stdout, "[3] Alice — Room 401 (double) — pending")
	assert.Contains(t, stdout, "Outgoing requests:")
	assert.Contains(t, stdout, "[4] Bob — accepted")
}

func TestRequestsCmd_Accept(t *testing.T) {
	var acceptedID int64
	cmd := newRequestsCmd(&mockRequestsClient{
		acceptFn: func(_ context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
			acceptedID = requestID
			return &client.RoommateRequestResponse{ID: requestID, Status: "accepted"}, nil
		},
	})

	stdout, stderr, err := executeRequestsCmd(t, cmd, []string{"accept", "42"})
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, int64(42), acceptedID)
	assert.Contains(t, stdout, "Accepted roommate request 42")
}

func TestRequestsCmd_Decline(t *testing.T) {
	var declinedID int64
	cmd := newRequestsCmd(&mockRequestsClient{
		declineFn: func(_ context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
			declinedID = requestID
			return &client.RoommateRequestResponse{ID: requestID, Status: "declined"}, nil
		},
	})

	stdout, stderr, err := executeRequestsCmd(t, cmd, []string{"decline", "43"})
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, int64(43), declinedID)
	assert.Contains(t, stdout, "Declined roommate request 43")
}

func TestRequestsCmd_Accept_InvalidRequestID(t *testing.T) {
	cmd := newRequestsCmd(&mockRequestsClient{})
	_, _, err := executeRequestsCmd(t, cmd, []string{"accept", "bad-id"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request_id must be a positive integer")
}

func TestRequestsCmd_ErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStderr string
	}{
		{name: "not authenticated", err: auth.ErrNotAuthenticated, wantStderr: "not logged in"},
		{name: "session expired", err: client.ErrSessionExpired, wantStderr: "session has expired"},
		{name: "unauthorized", err: client.ErrUnauthorized, wantStderr: "not authorized"},
		{name: "not found", err: client.ErrRequestNotFound, wantStderr: "not found"},
		{name: "forbidden", err: client.ErrRequestForbidden, wantStderr: "only respond"},
		{name: "invalid state", err: client.ErrInvalidRequestState, wantStderr: "already resolved"},
		{name: "generic", err: errors.New("boom"), wantStderr: "Could not accept request"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newRequestsCmd(&mockRequestsClient{
				acceptFn: func(_ context.Context, _ int64) (*client.RoommateRequestResponse, error) {
					return nil, tc.err
				},
			})
			_, stderr, err := executeRequestsCmd(t, cmd, []string{"accept", "11"})
			require.Error(t, err)
			assert.Contains(t, stderr, tc.wantStderr)
		})
	}
}
