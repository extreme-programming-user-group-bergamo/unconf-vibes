package main

import (
	"context"
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
	"github.com/katurdays/unconf/internal/config"
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

	listenAddr := resolveListenAddr(cfg.GetAPIEndpoint())
	router := api.NewRouter()

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
