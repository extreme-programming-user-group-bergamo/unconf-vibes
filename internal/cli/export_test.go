package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockExportClient struct {
	csv             []byte
	err             error
	slug            string
	includeCanceled bool
}

func (m *mockExportClient) ExportConferenceBookingsCSV(_ context.Context, slug string, includeCancelled bool) ([]byte, error) {
	m.slug = slug
	m.includeCanceled = includeCancelled
	if m.err != nil {
		return nil, m.err
	}
	return m.csv, nil
}

type mockExportContextStore struct {
	activeConference string
	err              error
}

func (m *mockExportContextStore) GetActiveConference() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.activeConference, nil
}

func TestExportCmd_WritesCSVToStdoutByDefault(t *testing.T) {
	mockClient := &mockExportClient{csv: []byte("name,email\nAlice,alice@test.dev\n")}
	cmd := newExportCmd(mockClient, &mockExportContextStore{activeConference: "socrates-26"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "socrates-26", mockClient.slug)
	assert.Equal(t, "name,email\nAlice,alice@test.dev\n", stdout.String())
}

func TestExportCmd_WritesCSVToFile(t *testing.T) {
	mockClient := &mockExportClient{csv: []byte("name,email\nAlice,alice@test.dev\n")}
	cmd := newExportCmd(mockClient, &mockExportContextStore{activeConference: "socrates-26"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})

	outputPath := filepath.Join(t.TempDir(), "bookings.csv")
	cmd.SetArgs([]string{"--output", outputPath})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Exported conference bookings CSV to")

	data, readErr := os.ReadFile(outputPath)
	require.NoError(t, readErr)
	assert.Equal(t, "name,email\nAlice,alice@test.dev\n", string(data))
}

func TestExportCmd_IncludeCancelledFlagAndArgSlug(t *testing.T) {
	mockClient := &mockExportClient{csv: []byte("name,email\n")}
	cmd := newExportCmd(mockClient, &mockExportContextStore{activeConference: "ignored"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"manual-conf", "--include-cancelled"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "manual-conf", mockClient.slug)
	assert.True(t, mockClient.includeCanceled)
}

func TestExportCmd_NoContextShowsMessage(t *testing.T) {
	cmd := newExportCmd(&mockExportClient{}, &mockExportContextStore{})
	cmd.SetOut(&bytes.Buffer{})
	var stderr bytes.Buffer
	cmd.SetErr(&stderr)

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stderr.String(), "No conference specified")
}

func TestExportCmd_ContextReadFailure(t *testing.T) {
	cmd := newExportCmd(&mockExportClient{}, &mockExportContextStore{err: errors.New("boom")})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read conference context")
}

func TestExportCmd_ErrorMappings(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantErr    error
		wantStderr string
	}{
		{name: "not authenticated", err: auth.ErrNotAuthenticated, wantErr: auth.ErrNotAuthenticated, wantStderr: "not logged in"},
		{name: "session expired", err: client.ErrSessionExpired, wantErr: client.ErrSessionExpired, wantStderr: "session has expired"},
		{name: "forbidden", err: client.ErrOrganizerForbidden, wantErr: client.ErrOrganizerForbidden, wantStderr: "Organizer permissions required"},
		{name: "conference missing", err: client.ErrConferenceNotFound, wantErr: client.ErrConferenceNotFound, wantStderr: "not found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newExportCmd(&mockExportClient{err: tc.err}, &mockExportContextStore{activeConference: "socrates-26"})
			cmd.SetOut(&bytes.Buffer{})
			var stderr bytes.Buffer
			cmd.SetErr(&stderr)

			err := cmd.Execute()
			require.Error(t, err)
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Contains(t, stderr.String(), tc.wantStderr)
		})
	}
}
