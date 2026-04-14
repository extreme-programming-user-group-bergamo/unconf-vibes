package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migrateSqlite "github.com/golang-migrate/migrate/v4/database/sqlite3"
	migrateIOFS "github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/katurdays/unconf/internal/config"
	"github.com/katurdays/unconf/internal/repository"
	appMigrations "github.com/katurdays/unconf/migrations"
	_ "github.com/mattn/go-sqlite3"
)

const defaultDBPath = ".unconf.db"

func NewConnectionManager(ctx context.Context, dbPath string) (*sql.DB, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled before opening sqlite connection: %w", err)
	}

	cfg, err := config.LoadConfig(ctx, config.LoadOptions{})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to load database configuration: %v", repository.ErrDatabaseInit, err)
	}

	maxOpenConns := cfg.GetDBMaxOpenConns()
	if maxOpenConns <= 0 {
		maxOpenConns = 25
	}

	maxIdleConns := cfg.GetDBMaxIdleConns()
	if maxIdleConns < 0 {
		maxIdleConns = 25
	}

	busyTimeoutMS := cfg.GetDBBusyTimeoutMS()
	if busyTimeoutMS < 0 {
		busyTimeoutMS = 5000
	}

	resolvedPath := dbPath
	if resolvedPath == "" {
		resolvedPath = cfg.GetDBPath()
		if resolvedPath == "" {
			resolvedPath = defaultDBPath
		}
	}

	connectionDSN := withSQLiteConnectionParams(resolvedPath, busyTimeoutMS)

	if resolvedPath != ":memory:" {
		if err := ensureDBDirExists(resolvedPath); err != nil {
			return nil, fmt.Errorf("%w: failed to ensure database directory exists: %v", repository.ErrDatabaseInit, err)
		}

		if err := ensureDBPathWritable(resolvedPath); err != nil {
			return nil, fmt.Errorf("%w: database path is not writable (%s): %v", repository.ErrDatabaseInit, resolvedPath, err)
		}
	}

	db, err := sql.Open("sqlite3", connectionDSN)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to open sqlite database: %v", repository.ErrDatabaseInit, err)
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: failed to ping sqlite database: %v", repository.ErrDatabaseInit, err)
	}

	if err := enableForeignKeys(ctx, db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("%w: failed to enable sqlite foreign keys: %v", repository.ErrDatabaseInit, err)
	}

	slog.Info("sqlite database connection initialized", "db_path", resolvedPath)
	slog.Debug("sqlite connection pool configured", "max_open_conns", maxOpenConns, "max_idle_conns", maxIdleConns, "busy_timeout_ms", busyTimeoutMS)

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}

	const maxAttempts = 5
	backoff := 200 * time.Millisecond

	slog.Info("running database migrations", "source", "embedded")
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = m.Up()
		if err == nil || errors.Is(err, migrate.ErrNoChange) {
			slog.Info("database migrations complete")
			return nil
		}

		if !isTransientMigrationLockError(err) || attempt == maxAttempts {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		slog.Warn("migration lock detected, retrying",
			"attempt", attempt,
			"max_attempts", maxAttempts,
			"backoff_ms", backoff.Milliseconds(),
			"error", err,
		)

		time.Sleep(backoff)
		backoff *= 2
	}

	return fmt.Errorf("failed to run migrations: exceeded retry attempts")
}

func MigrationStatus(db *sql.DB) (uint, bool, error) {
	m, err := newMigrator(db)
	if err != nil {
		return 0, false, err
	}

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("failed to fetch migration version: %w", err)
	}

	return version, dirty, nil
}

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database handle is nil", repository.ErrDatabaseInit)
	}

	driver, err := migrateSqlite.WithInstance(db, &migrateSqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("%w: failed to initialize sqlite migration driver: %v", repository.ErrDatabaseInit, err)
	}

	sourceDriver, err := migrateIOFS.New(appMigrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to initialize embedded migration source: %v", repository.ErrDatabaseInit, err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite3", driver)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to initialize migration runner: %v", repository.ErrDatabaseInit, err)
	}

	return m, nil
}

func ensureDBDirExists(dbPath string) error {
	cleanPath := filepath.Clean(dbPath)
	dir := filepath.Dir(cleanPath)
	if dir == "." {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create database directory %q: %w", dir, err)
	}

	return nil
}

func ensureDBPathWritable(dbPath string) error {
	file, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open database file for read/write: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close writable database file handle: %w", err)
	}

	return nil
}

func withBusyTimeout(dsn string, timeoutMS int) string {
	separator := "?"
	if containsQuery(dsn) {
		separator = "&"
	}

	return dsn + separator + "_busy_timeout=" + strconv.Itoa(timeoutMS)
}

func withSQLiteConnectionParams(dsn string, timeoutMS int) string {
	withTimeout := withBusyTimeout(dsn, timeoutMS)
	separator := "&"
	if !containsQuery(withTimeout) {
		separator = "?"
	}

	return withTimeout + separator + "_foreign_keys=on"
}

func enableForeignKeys(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("pragma enable failed: %w", err)
	}

	var foreignKeysEnabled int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeysEnabled); err != nil {
		return fmt.Errorf("pragma verify failed: %w", err)
	}

	if foreignKeysEnabled != 1 {
		return fmt.Errorf("pragma verify failed: foreign_keys=%d", foreignKeysEnabled)
	}

	return nil
}

func containsQuery(dsn string) bool {
	return strings.Contains(dsn, "?")
}

func isTransientMigrationLockError(err error) bool {
	if err == nil {
		return false
	}

	errMessage := strings.ToLower(err.Error())
	return strings.Contains(errMessage, "database is locked") || strings.Contains(errMessage, "database table is locked")
}
