package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/katurdays/unconf/internal/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEditConferenceClient struct {
	getConferenceFn    func(ctx context.Context, slug string) (*client.ConferenceResponse, error)
	updateConferenceFn func(ctx context.Context, slug string, input client.UpdateConferenceRequest) (*client.ConferenceResponse, error)
}

func (m *mockEditConferenceClient) GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error) {
	if m.getConferenceFn != nil {
		return m.getConferenceFn(ctx, slug)
	}
	return &client.ConferenceResponse{}, nil
}

func (m *mockEditConferenceClient) UpdateConference(ctx context.Context, slug string, input client.UpdateConferenceRequest) (*client.ConferenceResponse, error) {
	if m.updateConferenceFn != nil {
		return m.updateConferenceFn(ctx, slug, input)
	}
	return &client.ConferenceResponse{Slug: slug}, nil
}

type mockEditContextStore struct {
	getActiveConferenceFn func() (string, error)
}

func (m *mockEditContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

func executeEditCmd(t *testing.T, cmd *cobra.Command, args []string, stdin string) (string, string, error) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(stdin))
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestEditCmd_UsesContextAndUpdatesConference(t *testing.T) {
	var captured client.UpdateConferenceRequest
	mockClient := &mockEditConferenceClient{
		getConferenceFn: func(_ context.Context, slug string) (*client.ConferenceResponse, error) {
			assert.Equal(t, "ctx-conf", slug)
			return &client.ConferenceResponse{
				Slug: "ctx-conf", Name: "Before", Description: "desc", Location: "Berlin",
				StartDate: "2026-10-07", EndDate: "2026-10-10", Capacity: 100,
			}, nil
		},
		updateConferenceFn: func(_ context.Context, slug string, input client.UpdateConferenceRequest) (*client.ConferenceResponse, error) {
			assert.Equal(t, "ctx-conf", slug)
			captured = input
			return &client.ConferenceResponse{Slug: slug}, nil
		},
	}
	ctxStore := &mockEditContextStore{getActiveConferenceFn: func() (string, error) { return "ctx-conf", nil }}

	cmd := newEditCmd(mockClient, ctxStore)
	stdout, stderr, err := executeEditCmd(t, cmd, nil, "After\ndesc2\nMunich\n2026-10-08\n2026-10-11\n120\n")
	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, "After", captured.Name)
	assert.Equal(t, "Munich", captured.Location)
	assert.Contains(t, stdout, "Conference updated successfully: ctx-conf")
}

func TestEditCmd_NoContextOrArg_PrintsGuidance(t *testing.T) {
	cmd := newEditCmd(&mockEditConferenceClient{}, &mockEditContextStore{})
	_, stderr, err := executeEditCmd(t, cmd, nil, "")
	require.NoError(t, err)
	assert.Contains(t, stderr, "No conference specified")
}
