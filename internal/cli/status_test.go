package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStatusContextStore struct {
	activeConference string
	err              error
}

func (m *mockStatusContextStore) GetActiveConference() (string, error) {
	if m.err != nil {
		return "", m.err
	}

	return m.activeConference, nil
}

func TestStatusCmd_UsesActiveConferenceScope(t *testing.T) {
	cmd := newStatusCmd(struct{}{}, &mockStatusContextStore{activeConference: "socrates-2026"})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "active conference \"socrates-2026\"")
}

func TestStatusCmd_AllFlagBypassesContextRequirement(t *testing.T) {
	cmd := newStatusCmd(struct{}{}, &mockStatusContextStore{activeConference: ""})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "across all conferences")
}

func TestStatusCmd_NoActiveConferenceMessage(t *testing.T) {
	cmd := newStatusCmd(struct{}{}, &mockStatusContextStore{})
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "No active conference context")
}

func TestStatusCmd_ContextReadFailure(t *testing.T) {
	cmd := newStatusCmd(struct{}{}, &mockStatusContextStore{err: errors.New("boom")})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read conference context")
}
