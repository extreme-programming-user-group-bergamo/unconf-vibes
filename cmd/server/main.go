package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/katurdays/unconf/internal/api"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/config"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/katurdays/unconf/internal/service"
)

const shutdownTimeout = 30 * time.Second

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	cfg, err := config.LoadConfig(ctx, config.LoadOptions{})
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	if strings.TrimSpace(cfg.GetGitHubClientID()) == "" || strings.TrimSpace(cfg.GetGitHubClientSecret()) == "" {
		slog.Error("failed to initialize github oauth provider", "error", "github_client_id and github_client_secret must be set")
		os.Exit(1)
	}

	listenAddr := resolveListenAddr(cfg.GetAPIEndpoint())

	db, err := initializeDatabase(ctx, cfg)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			slog.Error("failed to close database connection", "error", closeErr)
		}
	}()

	userRepository := sqlite.NewUserRepository(db)
	refreshSessionRepository := sqlite.NewRefreshSessionRepository(db)

	githubProvider := auth.NewHTTPGitHubProvider(cfg.GetGitHubClientID(), cfg.GetGitHubClientSecret(), 10*time.Second)
	tokenService, err := auth.NewTokenService(cfg.GetPasetoSymmetricKey())
	if err != nil {
		slog.Error("failed to initialize token service", "error", err)
		os.Exit(1)
	}

	authService, err := service.NewAuthService(
		githubProvider,
		tokenService,
		userRepository,
		refreshSessionRepository,
		cfg.GetTokenTTL(),
		cfg.GetRefreshTTL(),
	)
	if err != nil {
		slog.Error("failed to initialize auth service", "error", err)
		os.Exit(1)
	}

	authHandler := handlers.NewAuthHandler(authService)

	userService := service.NewUserService(userRepository)
	userHandler := handlers.NewUserHandler(userService)

	conferenceRepo := sqlite.NewConferenceRepository(db)
	bookingRepo := sqlite.NewBookingRepository(db)
	conferenceService := service.NewConferenceService(conferenceRepo, bookingRepo)
	conferenceHandler := handlers.NewConferenceHandler(conferenceService)

	router := api.NewRouter(authHandler, tokenService, userHandler, conferenceHandler)

	server := &http.Server{
		Addr:    listenAddr,
		Handler: router,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.Info("UNCONF API starting", "addr", listenAddr)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errChan <- fmt.Errorf("failed to start server: %w", serveErr)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	case sig := <-sigChan:
		slog.Info("graceful shutdown initiated", "signal", sig.String())

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shutdown server gracefully", "error", err)
			os.Exit(1)
		}

		slog.Info("server stopped cleanly")
		os.Exit(0)
	}
}

func initializeDatabase(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	db, err := sqlite.NewConnectionManager(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite connection manager: %w", err)
	}

	if err := sqlite.RunMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	version, dirty, err := sqlite.MigrationStatus(db)
	if err != nil {
		slog.Warn("failed to read database migration status", "error", err)
	} else {
		slog.Info("database startup diagnostics",
			"db_path", cfg.GetDBPath(),
			"db_max_open_conns", cfg.GetDBMaxOpenConns(),
			"db_max_idle_conns", cfg.GetDBMaxIdleConns(),
			"db_busy_timeout_ms", cfg.GetDBBusyTimeoutMS(),
			"migration_version", version,
			"migration_dirty", dirty,
		)
	}

	slog.Info("database initialization complete")
	return db, nil
}

func resolveListenAddr(apiEndpoint string) string {
	const defaultAddr = ":8080"

	if apiEndpoint == "" {
		return defaultAddr
	}

	if strings.Contains(apiEndpoint, "://") {
		parsed, err := url.Parse(apiEndpoint)
		if err != nil {
			return defaultAddr
		}

		host := parsed.Host
		if host == "" {
			return defaultAddr
		}

		if _, port, err := net.SplitHostPort(host); err == nil {
			return ":" + port
		}

		return defaultAddr
	}

	if strings.HasPrefix(apiEndpoint, ":") {
		return apiEndpoint
	}

	if _, port, err := net.SplitHostPort(apiEndpoint); err == nil {
		return ":" + port
	}

	return defaultAddr
}
