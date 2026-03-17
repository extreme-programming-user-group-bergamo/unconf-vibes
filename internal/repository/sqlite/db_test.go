package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionManager_AppliesPoolSettingsFromConfig(t *testing.T) {
	t.Setenv("UNCONF_DB_MAX_OPEN_CONNS", "7")
	t.Setenv("UNCONF_DB_MAX_IDLE_CONNS", "4")

	dbPath := filepath.Join(t.TempDir(), "pool.db")
	db, err := NewConnectionManager(context.Background(), dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	stats := db.Stats()
	assert.Equal(t, 7, stats.MaxOpenConnections)
}

func TestWithBusyTimeout(t *testing.T) {
	assert.Equal(t, "test.db?_busy_timeout=5000", withBusyTimeout("test.db", 5000))
	assert.Equal(t, "test.db?cache=shared&_busy_timeout=5000", withBusyTimeout("test.db?cache=shared", 5000))
}

func TestIsTransientMigrationLockError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{name: "nil", err: nil, expected: false},
		{name: "database locked", err: errors.New("database is locked"), expected: true},
		{name: "table locked", err: errors.New("database table is locked: schema_migrations"), expected: true},
		{name: "other error", err: errors.New("syntax error"), expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, isTransientMigrationLockError(testCase.err))
		})
	}
}

func TestNewConnectionManager_UnwritableDatabasePath(t *testing.T) {
	dir := t.TempDir()
	readOnlyDir := filepath.Join(dir, "readonly")
	require.NoError(t, os.MkdirAll(readOnlyDir, 0o755))
	require.NoError(t, os.Chmod(readOnlyDir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(readOnlyDir, 0o755)
	})

	dbPath := filepath.Join(readOnlyDir, "blocked.db")
	_, err := NewConnectionManager(context.Background(), dbPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database path is not writable")
}
