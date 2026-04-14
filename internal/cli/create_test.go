package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateConferenceClient struct {
	createConferenceFn func(ctx context.Context, input client.CreateConferenceRequest) (*client.ConferenceResponse, error)
}

func (m *mockCreateConferenceClient) CreateConference(ctx context.Context, input client.CreateConferenceRequest) (*client.ConferenceResponse, error) {
	if m.createConferenceFn != nil {
		return m.createConferenceFn(ctx, input)
	}
	return &client.ConferenceResponse{Slug: input.Slug}, nil
}

func executeCreateCmd(t *testing.T, cmd *cobra.Command, stdin string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetIn(strings.NewReader(stdin))
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestCreateCmd_InteractiveSuccess(t *testing.T) {
	var captured client.CreateConferenceRequest
	mockClient := &mockCreateConferenceClient{
		createConferenceFn: func(_ context.Context, input client.CreateConferenceRequest) (*client.ConferenceResponse, error) {
			captured = input
			return &client.ConferenceResponse{Slug: input.Slug}, nil
		},
	}

	cmd := newCreateCmd(mockClient)
	stdin := "SoCraTes 2026\nSoCraTes_26\nConference description\nBerlin\n2026-10-07\n2026-10-10\n120\n"
	stdout, stderr, err := executeCreateCmd(t, cmd, stdin)
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, "socrates-26", captured.Slug)
	assert.Equal(t, 120, captured.Capacity)
	assert.Contains(t, stdout, "Conference created successfully: socrates-26")
}

func TestCreateCmd_InvalidDate(t *testing.T) {
	cmd := newCreateCmd(&mockCreateConferenceClient{})
	stdin := "Name\nslug\nDesc\nBerlin\nbad-date\n2026-10-10\n100\n"
	_, _, err := executeCreateCmd(t, cmd, stdin)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid start date")
}

func TestCreateCmd_ConferenceExists(t *testing.T) {
	mockClient := &mockCreateConferenceClient{
		createConferenceFn: func(_ context.Context, _ client.CreateConferenceRequest) (*client.ConferenceResponse, error) {
			return nil, fmt.Errorf("conflict: %w", client.ErrConferenceExists)
		},
	}
	cmd := newCreateCmd(mockClient)
	stdin := "Name\nslug\nDesc\nBerlin\n2026-10-07\n2026-10-10\n100\n"
	_, stderr, err := executeCreateCmd(t, cmd, stdin)
	require.Error(t, err)
	assert.Contains(t, stderr, "Conference slug already exists")
}
