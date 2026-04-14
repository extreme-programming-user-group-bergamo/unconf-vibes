package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/katurdays/unconf/internal/repository/sqlite"
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

func TestRootCmdIncludesDBCommand(t *testing.T) {
	cmd := NewRootCmd()

	dbCmd, _, err := cmd.Find([]string{"db", "status"})
	require.NoError(t, err)
	require.NotNil(t, dbCmd)
	assert.Equal(t, "status", dbCmd.Name())
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

func TestDBStatusCommand(t *testing.T) {
	t.Setenv("UNCONF_API_ENDPOINT", "http://127.0.0.1:8080")
	t.Setenv("UNCONF_LOG_LEVEL", "debug")
	t.Setenv("UNCONF_DB_PATH", filepath.Join(t.TempDir(), "status.db"))

	ctx := context.Background()
	db, err := sqlite.NewConnectionManager(ctx, "")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})
	require.NoError(t, sqlite.RunMigrations(db))

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"db", "status"})

	err = cmd.Execute()
	require.NoError(t, err)

	version, _, err := sqlite.MigrationStatus(db)
	require.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "db_path=")
	assert.Contains(t, output, "migration_version="+strconv.FormatUint(uint64(version), 10))
	assert.Contains(t, output, "dirty=false")
}
