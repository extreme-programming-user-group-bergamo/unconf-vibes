package servercli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "v1.0.0-alpha"

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "unconf-server",
		Short: "UNCONF Server CLI — Backend and admin operations",
		Long: `UNCONF Server CLI provides backend and operational commands.

Use this command to run API lifecycle operations and backend/admin utilities.`,
		Version:      Version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runServe(cmd.Context())
		},
	}

	rootCmd.SetVersionTemplate("{{printf \"%s %s\\n\" .Name .Version}}")
	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newDBCmd())

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
