package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/config"
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

			slog.Info("UNCONF CLI starting",
				"version", Version,
				"config_file", cfg.GetConfigFile(),
				"command", cmd.CommandPath(),
			)

			return nil
		},
	}

	rootCmd.SetVersionTemplate("{{printf \"%s %s\\n\" .Name .Version}}")
	rootCmd.Flags().StringVar(&configFile, "config", "", "Path to config file (default: ./.unconf.yaml or ~/.unconf.yaml)")

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
