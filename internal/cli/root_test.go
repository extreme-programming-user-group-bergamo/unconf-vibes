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

func TestRootCmdDoesNotIncludeDBCommand(t *testing.T) {
	cmd := NewRootCmd()

	dbCmd, _, err := cmd.Find([]string{"db", "status"})
	require.ErrorContains(t, err, "unknown command \"db\" for \"unconf\"")
	require.NotNil(t, dbCmd)
	assert.Equal(t, "unconf", dbCmd.Name())
}

func TestRootCmdIncludesRoomsCommand(t *testing.T) {
	cmd := NewRootCmd()

	roomsCmd, _, err := cmd.Find([]string{"rooms"})
	require.NoError(t, err)
	require.NotNil(t, roomsCmd)
	assert.Equal(t, "rooms", roomsCmd.Name())
}

func TestRootCmdIncludesAttendeesCommand(t *testing.T) {
	cmd := NewRootCmd()

	attendeesCmd, _, err := cmd.Find([]string{"attendees"})
	require.NoError(t, err)
	require.NotNil(t, attendeesCmd)
	assert.Equal(t, "attendees", attendeesCmd.Name())
}

func TestRootCmdIncludesDashboardCommand(t *testing.T) {
	cmd := NewRootCmd()

	dashboardCmd, _, err := cmd.Find([]string{"dashboard"})
	require.NoError(t, err)
	require.NotNil(t, dashboardCmd)
	assert.Equal(t, "dashboard", dashboardCmd.Name())
}

func TestRootCmdIncludesInviteCommand(t *testing.T) {
	cmd := NewRootCmd()

	inviteCmd, _, err := cmd.Find([]string{"invite"})
	require.NoError(t, err)
	require.NotNil(t, inviteCmd)
	assert.Equal(t, "invite", inviteCmd.Name())
}

func TestRootCmdIncludesRequestsCommand(t *testing.T) {
	cmd := NewRootCmd()

	requestsCmd, _, err := cmd.Find([]string{"requests"})
	require.NoError(t, err)
	require.NotNil(t, requestsCmd)
	assert.Equal(t, "requests", requestsCmd.Name())
}

func TestRootCmdIncludesCancelCommand(t *testing.T) {
	cmd := NewRootCmd()

	cancelCmd, _, err := cmd.Find([]string{"cancel"})
	require.NoError(t, err)
	require.NotNil(t, cancelCmd)
	assert.Equal(t, "cancel", cancelCmd.Name())
}

func TestRootCmdIncludesCreateCommand(t *testing.T) {
	cmd := NewRootCmd()

	createCmd, _, err := cmd.Find([]string{"create"})
	require.NoError(t, err)
	require.NotNil(t, createCmd)
	assert.Equal(t, "create", createCmd.Name())
}

func TestRootCmdIncludesEditCommand(t *testing.T) {
	cmd := NewRootCmd()

	editCmd, _, err := cmd.Find([]string{"edit"})
	require.NoError(t, err)
	require.NotNil(t, editCmd)
	assert.Equal(t, "edit", editCmd.Name())
}

func TestRootCmdIncludesExportCommand(t *testing.T) {
	cmd := NewRootCmd()

	exportCmd, _, err := cmd.Find([]string{"export"})
	require.NoError(t, err)
	require.NotNil(t, exportCmd)
	assert.Equal(t, "export", exportCmd.Name())
}

func TestRootCmdDoesNotIncludeConfirmCommand(t *testing.T) {
	cmd := NewRootCmd()

	confirmCmd, _, err := cmd.Find([]string{"confirm"})
	require.ErrorContains(t, err, "unknown command \"confirm\" for \"unconf\"")
	require.NotNil(t, confirmCmd)
	assert.Equal(t, "unconf", confirmCmd.Name())
}
