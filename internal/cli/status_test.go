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
	getMeFn func(ctx context.Context, accessToken string) (*client.UserResponse, error)
}

func (m *mockStatusClient) GetMe(ctx context.Context, accessToken string) (*client.UserResponse, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx, accessToken)
	}

	return nil, nil
}

func TestStatusCmd_AuthenticatedUser(t *testing.T) {
	store := auth.NewMockTokenStore()
	store.SetTokens("valid-access-token", "valid-refresh-token")

	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context, token string) (*client.UserResponse, error) {
			assert.Equal(t, "valid-access-token", token)
			return &client.UserResponse{
				ID:          42,
				GitHubID:    "12345",
				Email:       "user@example.com",
				DisplayName: "octocat",
			}, nil
		},
	}

	cmd := newStatusCmd(statusClient, store)
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
	store := auth.NewMockTokenStore()

	statusClient := &mockStatusClient{}

	cmd := newStatusCmd(statusClient, store)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
	assert.Contains(t, stderr.String(), "not logged in")
}

func TestStatusCmd_API401ClearsTokens(t *testing.T) {
	store := auth.NewMockTokenStore()
	store.SetTokens("expired-token", "refresh-token")

	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context, _ string) (*client.UserResponse, error) {
			return nil, client.ErrUnauthorized
		},
	}

	cmd := newStatusCmd(statusClient, store)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrUnauthorized)
	assert.Contains(t, stderr.String(), "session has expired")

	// Verify tokens were cleared
	assert.False(t, store.HasValidToken())
}

func TestStatusCmd_NetworkError(t *testing.T) {
	store := auth.NewMockTokenStore()
	store.SetTokens("valid-token", "refresh-token")

	statusClient := &mockStatusClient{
		getMeFn: func(_ context.Context, _ string) (*client.UserResponse, error) {
			return nil, assert.AnError
		},
	}

	cmd := newStatusCmd(statusClient, store)
	var stderr bytes.Buffer
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "Could not connect")
}
