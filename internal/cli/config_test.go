package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfigClient struct {
	getMeFn    func(ctx context.Context) (*client.UserResponse, error)
	updateMeFn func(ctx context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error)
}

func (m *mockConfigClient) GetMe(ctx context.Context) (*client.UserResponse, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx)
	}
	return nil, nil
}

func (m *mockConfigClient) UpdateMe(ctx context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error) {
	if m.updateMeFn != nil {
		return m.updateMeFn(ctx, input)
	}
	return nil, nil
}

func TestConfigCmd_Show_HappyPath(t *testing.T) {
	mock := &mockConfigClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return &client.UserResponse{
				ID:             42,
				GitHubID:       "12345",
				Email:          "jane@example.com",
				DisplayName:    "Jane Doe",
				PrivacySetting: "public",
			}, nil
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--show"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Profile Settings")
	assert.Contains(t, output, "Jane Doe")
	assert.Contains(t, output, "jane@example.com")
	assert.Contains(t, output, "public")
}

func TestConfigCmd_Show_NotAuthenticated(t *testing.T) {
	mock := &mockConfigClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, auth.ErrNotAuthenticated
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--show"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
	assert.Contains(t, stderr.String(), "not logged in")
}

func TestConfigCmd_Show_SessionExpired(t *testing.T) {
	mock := &mockConfigClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, fmt.Errorf("wrapped: %w", client.ErrSessionExpired)
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--show"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, client.ErrSessionExpired)
	assert.Contains(t, stderr.String(), "session has expired")
}

func TestConfigCmd_Show_ServerError(t *testing.T) {
	mock := &mockConfigClient{
		getMeFn: func(_ context.Context) (*client.UserResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--show"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get profile")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestConfigCmd_UpdateDisplayName_HappyPath(t *testing.T) {
	var capturedInput client.UpdateProfileRequest
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error) {
			capturedInput = input
			return &client.UserResponse{
				ID:             42,
				DisplayName:    "New Name",
				PrivacySetting: "public",
			}, nil
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--display-name", "New Name"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Profile updated successfully.")
	require.NotNil(t, capturedInput.DisplayName)
	assert.Equal(t, "New Name", *capturedInput.DisplayName)
	assert.Nil(t, capturedInput.PrivacySetting)
}

func TestConfigCmd_UpdatePrivacy_HappyPath(t *testing.T) {
	var capturedInput client.UpdateProfileRequest
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error) {
			capturedInput = input
			return &client.UserResponse{
				ID:             42,
				DisplayName:    "Jane",
				PrivacySetting: "private",
			}, nil
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--privacy", "private"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Profile updated successfully.")
	require.NotNil(t, capturedInput.PrivacySetting)
	assert.Equal(t, "private", *capturedInput.PrivacySetting)
	assert.Nil(t, capturedInput.DisplayName)
}

func TestConfigCmd_UpdateBothFlags(t *testing.T) {
	var capturedInput client.UpdateProfileRequest
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, input client.UpdateProfileRequest) (*client.UserResponse, error) {
			capturedInput = input
			return &client.UserResponse{
				ID:             42,
				DisplayName:    "New Name",
				PrivacySetting: "private",
			}, nil
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--display-name", "New Name", "--privacy", "private"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Profile updated successfully.")
	require.NotNil(t, capturedInput.DisplayName)
	require.NotNil(t, capturedInput.PrivacySetting)
	assert.Equal(t, "New Name", *capturedInput.DisplayName)
	assert.Equal(t, "private", *capturedInput.PrivacySetting)
}

func TestConfigCmd_UpdateInvalidPrivacy(t *testing.T) {
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, _ client.UpdateProfileRequest) (*client.UserResponse, error) {
			t.Fatal("UpdateMe should not be called for invalid privacy setting")
			return nil, nil
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--privacy", "invalid"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, stderr.String(), "Invalid privacy setting")
	assert.Contains(t, err.Error(), "invalid privacy setting")
}

func TestConfigCmd_Update_NotAuthenticated(t *testing.T) {
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, _ client.UpdateProfileRequest) (*client.UserResponse, error) {
			return nil, auth.ErrNotAuthenticated
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--display-name", "New Name"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.ErrorIs(t, err, auth.ErrNotAuthenticated)
	assert.Contains(t, stderr.String(), "not logged in")
}

func TestConfigCmd_Update_ServerError(t *testing.T) {
	mock := &mockConfigClient{
		updateMeFn: func(_ context.Context, _ client.UpdateProfileRequest) (*client.UserResponse, error) {
			return nil, fmt.Errorf("internal server error")
		},
	}

	cmd := newConfigCmd(mock)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"--display-name", "New Name"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update profile")
	assert.Contains(t, err.Error(), "internal server error")
}
