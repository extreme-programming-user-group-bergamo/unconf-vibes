package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatusClient struct {
	getMeFn func(ctx context.Context) (*client.UserResponse, error)
}

func (m *mockStatusClient) GetMe(ctx context.Context) (*client.UserResponse, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx)
	}

	return nil, nil
}

func TestStatusCmd_AuthenticatedUser(t *testing.T) {
	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return &client.UserResponse{
				ID:          42,
				GitHubID:    "12345",
				Email:       "user@example.com",
				DisplayName: "octocat",
			}, nil
		},
	}

	cmd := newStatusCmd(statusClient)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Logged in as octocat")
	assert.Contains(t, output, "user@example.com")
	assert.Contains(t, output, "12345")
}

func TestStatusCmd_NotLoggedIn(t *testing.T) {
	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, auth.ErrNotAuthenticated
		},
	}

	cmd := newStatusCmd(statusClient)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
	assert.Contains(t, stderr.String(), "not logged in")
}

func TestStatusCmd_SessionExpired(t *testing.T) {
	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, client.ErrSessionExpired
		},
	}

	cmd := newStatusCmd(statusClient)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrSessionExpired)
	assert.Contains(t, stderr.String(), "session has expired")
}

func TestStatusCmd_NetworkError(t *testing.T) {
	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := newStatusCmd(statusClient)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "Could not connect")
}
