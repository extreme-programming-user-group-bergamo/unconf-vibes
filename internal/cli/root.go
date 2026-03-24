package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/config"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

var Version = "v1.0.0-alpha"

func NewRootCmd() *cobra.Command {
	var configFile string

	rootCmd := &cobra.Command{
		Use:   "unconf",
		Short: "UNCONF CLI — Unconference registration made easy",
		Long: `UNCONF CLI helps attendees and organizers manage unconference events.

Use this command as the starting point for authentication, conference discovery,
room browsing, and booking workflows.`,
		Version:      Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(cmd.Context(), config.LoadOptions{
				ConfigFile: configFile,
				FlagSet:    cmd.Flags(),
				FlagBindings: map[string]string{
					"config_file": "config",
				},
			})
			if err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}

			db, err := sqlite.NewConnectionManager(cmd.Context(), "")
			if err != nil {
				return fmt.Errorf("failed to initialize database connection: %w", err)
			}
			defer func() {
				if closeErr := db.Close(); closeErr != nil {
					slog.Error("failed to close database connection", "error", closeErr)
				}
			}()

			if err := sqlite.RunMigrations(db); err != nil {
				return fmt.Errorf("failed to run database migrations: %w", err)
			}

			version, dirty, err := sqlite.MigrationStatus(db)
			if err != nil {
				slog.Warn("failed to read migration status", "error", err)
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

			slog.Info("UNCONF CLI starting",
				"version", Version,
				"config_file", cfg.GetConfigFile(),
				"db_path", cfg.GetDBPath(),
				"command", cmd.CommandPath(),
			)

			return nil
		},
	}

	rootCmd.SetVersionTemplate("{{printf \"%s %s\\n\" .Name .Version}}")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to config file (default: ./.unconf.yaml or ~/.unconf.yaml)")

	store := auth.NewKeyringTokenStore("unconf")

	cfg, cfgErr := config.LoadConfig(context.Background(), config.LoadOptions{
		ConfigFile: configFile,
	})
	apiEndpoint := "http://localhost:8080"
	if cfgErr == nil {
		apiEndpoint = cfg.GetAPIEndpoint()
	}
	apiClient := client.NewClient(apiEndpoint)
	authClient := client.NewAuthenticatedClient(apiClient, store)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to user config directory
		homeDir, err = os.UserConfigDir()
		if err != nil {
			slog.Warn("failed to determine home directory for context management", "error", err)
			homeDir = "." // Last resort: use current directory
		}
	}
	ctxManager := config.NewContextManager(filepath.Join(homeDir, ".unconf"))

	rootCmd.AddCommand(newDBCmd())
	rootCmd.AddCommand(newLoginCmd(apiClient, store))
	rootCmd.AddCommand(newLogoutCmd(apiClient, store))
	rootCmd.AddCommand(newStatusCmd(authClient))
	rootCmd.AddCommand(newListCmd(apiClient))
	rootCmd.AddCommand(newInfoCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newCheckoutCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newRoomsCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newConfigCmd(authClient))

	return rootCmd
}

func Execute(ctx context.Context) error {
	rootCmd := NewRootCmd()
	rootCmd.SetContext(ctx)

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute root command: %w", err)
	}

	return nil
}
