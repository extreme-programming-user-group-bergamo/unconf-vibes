package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	require.NotNil(t, cmd)
	assert.Equal(t, "unconf", cmd.Use)
	assert.Contains(t, cmd.Short, "UNCONF CLI")
}

func TestRootCmdHelpFlag(t *testing.T) {
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	helpOutput := out.String()
	assert.Contains(t, helpOutput, "UNCONF CLI helps attendees and organizers manage unconference events")
	assert.Contains(t, helpOutput, "Use this command as the starting point")
}

func TestRootCmdVersionFlag(t *testing.T) {
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	require.NoError(t, err)

	versionOutput := out.String()
	assert.Contains(t, versionOutput, Version)
	assert.Contains(t, versionOutput, "unconf")
}

func TestRootCmdInvalidFlagUsageError(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--definitely-not-a-real-flag"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Equal(t, UsageError, ExitCodeForError(err))
}
