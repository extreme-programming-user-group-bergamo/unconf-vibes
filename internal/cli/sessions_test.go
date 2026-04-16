package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionsClient struct {
	listFn         func(ctx context.Context) ([]client.SessionResponse, error)
	revokeFn       func(ctx context.Context, sessionID int64) error
	revokeOthersFn func(ctx context.Context) (int64, error)
}

func (m *mockSessionsClient) ListSessions(ctx context.Context) ([]client.SessionResponse, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockSessionsClient) RevokeSession(ctx context.Context, sessionID int64) error {
	if m.revokeFn != nil {
		return m.revokeFn(ctx, sessionID)
	}
	return nil
}

func (m *mockSessionsClient) RevokeOtherSessions(ctx context.Context) (int64, error) {
	if m.revokeOthersFn != nil {
		return m.revokeOthersFn(ctx)
	}
	return 0, nil
}

func TestSessionsListCommand(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		listFn: func(_ context.Context) ([]client.SessionResponse, error) {
			now := time.Now().UTC()
			return []client.SessionResponse{
				{ID: 7, Current: true, ClientMetadata: "ua=test", CreatedAt: now, LastSeenAt: now},
			}, nil
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Active Sessions")
	assert.Contains(t, stdout.String(), "ID: 7")
	assert.Contains(t, stdout.String(), "(current)")
}

func TestSessionsRevokeOthersCancelled(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		revokeOthersFn: func(_ context.Context) (int64, error) {
			t.Fatal("revoke others should not be called when confirmation is denied")
			return 0, nil
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(bytes.NewBufferString("n\n"))
	cmd.SetArgs([]string{"revoke-others"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Cancelled.")
}

func TestSessionsRevokeCommandCurrentSessionError(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		revokeFn: func(_ context.Context, _ int64) error {
			return fmt.Errorf("wrapped: %w", client.ErrCurrentSessionRevoke)
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"revoke", "9"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrCurrentSessionRevoke)
	assert.Contains(t, stderr.String(), "Cannot revoke your current session")
}

func TestSessionsRevokeCommandSuccess(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		revokeFn: func(_ context.Context, sessionID int64) error {
			assert.Equal(t, int64(9), sessionID)
			return nil
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"revoke", "9"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Session 9 revoked.")
	assert.Empty(t, stderr.String())
}

func TestSessionsRevokeOthersYesSuccess(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		revokeOthersFn: func(_ context.Context) (int64, error) {
			return 2, nil
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"revoke-others", "--yes"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Revoked 2 other session(s).")
	assert.Empty(t, stderr.String())
}

func TestSessionsListCommandEmptyState(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		listFn: func(_ context.Context) ([]client.SessionResponse, error) {
			return []client.SessionResponse{}, nil
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No active sessions found.")
	assert.Empty(t, stderr.String())
}

func TestSessionsListCommandNotAuthenticated(t *testing.T) {
	cmd := newSessionsCmd(&mockSessionsClient{
		listFn: func(_ context.Context) ([]client.SessionResponse, error) {
			return nil, fmt.Errorf("wrapped: %w", auth.ErrNotAuthenticated)
		},
	})

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, errors.Is(err, auth.ErrNotAuthenticated))
	assert.Contains(t, stderr.String(), "You are not logged in. Run 'unconf login' to authenticate.")
}
