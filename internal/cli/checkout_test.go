package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCheckoutClient struct {
	getConferenceFn func(ctx context.Context, slug string) (*client.ConferenceResponse, error)
}

func (m *mockCheckoutClient) GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error) {
	if m.getConferenceFn != nil {
		return m.getConferenceFn(ctx, slug)
	}
	return nil, nil
}

type mockContextStore struct {
	getActiveConferenceFn func() (string, error)
	setActiveConferenceFn func(slug string) error
}

func (m *mockContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

func (m *mockContextStore) SetActiveConference(slug string) error {
	if m.setActiveConferenceFn != nil {
		return m.setActiveConferenceFn(slug)
	}
	return nil
}

func TestCheckoutCmd_SetHappyPath(t *testing.T) {
	var savedSlug string
	mockClient := &mockCheckoutClient{
		getConferenceFn: func(_ context.Context, slug string) (*client.ConferenceResponse, error) {
			return &client.ConferenceResponse{Slug: slug, Name: "SoCraTes 2026"}, nil
		},
	}
	mockStore := &mockContextStore{
		setActiveConferenceFn: func(slug string) error {
			savedSlug = slug
			return nil
		},
	}

	cmd := newCheckoutCmd(mockClient, mockStore)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Switched to socrates-26")
	assert.Equal(t, "socrates-26", savedSlug)
}

func TestCheckoutCmd_SetNotFound(t *testing.T) {
	mockClient := &mockCheckoutClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return nil, client.ErrConferenceNotFound
		},
	}
	mockStore := &mockContextStore{}

	cmd := newCheckoutCmd(mockClient, mockStore)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to checkout")
	assert.Contains(t, stderr.String(), "not found")
	assert.Contains(t, stderr.String(), "unconf list")
}

func TestCheckoutCmd_SetAPIError(t *testing.T) {
	mockClient := &mockCheckoutClient{
		getConferenceFn: func(_ context.Context, _ string) (*client.ConferenceResponse, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	mockStore := &mockContextStore{}

	cmd := newCheckoutCmd(mockClient, mockStore)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate conference")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestCheckoutCmd_SetSaveError(t *testing.T) {
	mockClient := &mockCheckoutClient{
		getConferenceFn: func(_ context.Context, slug string) (*client.ConferenceResponse, error) {
			return &client.ConferenceResponse{Slug: slug}, nil
		},
	}
	mockStore := &mockContextStore{
		setActiveConferenceFn: func(_ string) error {
			return fmt.Errorf("disk full")
		},
	}

	cmd := newCheckoutCmd(mockClient, mockStore)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"socrates-26"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save conference context")
	assert.Contains(t, err.Error(), "disk full")
}

func TestCheckoutCmd_SetVerifiesSlug(t *testing.T) {
	var receivedSlug string
	mockClient := &mockCheckoutClient{
		getConferenceFn: func(_ context.Context, slug string) (*client.ConferenceResponse, error) {
			receivedSlug = slug
			return &client.ConferenceResponse{Slug: slug}, nil
		},
	}
	mockStore := &mockContextStore{}

	cmd := newCheckoutCmd(mockClient, mockStore)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"my-conf-2026"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "my-conf-2026", receivedSlug)
}

func TestCheckoutCmd_ShowWithContext(t *testing.T) {
	mockClient := &mockCheckoutClient{}
	mockStore := &mockContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "socrates-26", nil
		},
	}

	cmd := newCheckoutCmd(mockClient, mockStore)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Current conference: socrates-26")
}

func TestCheckoutCmd_ShowNoContext(t *testing.T) {
	mockClient := &mockCheckoutClient{}
	mockStore := &mockContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "", nil
		},
	}

	cmd := newCheckoutCmd(mockClient, mockStore)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No conference selected")
	assert.Contains(t, stdout.String(), "unconf checkout <slug>")
}

func TestCheckoutCmd_ShowReadError(t *testing.T) {
	mockClient := &mockCheckoutClient{}
	mockStore := &mockContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "", fmt.Errorf("permission denied")
		},
	}

	cmd := newCheckoutCmd(mockClient, mockStore)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read current context")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestCheckoutCmd_TooManyArgs(t *testing.T) {
	mockClient := &mockCheckoutClient{}
	mockStore := &mockContextStore{}

	cmd := newCheckoutCmd(mockClient, mockStore)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"slug1", "slug2"})

	err := cmd.Execute()
	require.Error(t, err)
}
