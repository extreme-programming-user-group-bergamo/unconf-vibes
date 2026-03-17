package cli

import (
	"fmt"

	"github.com/katurdays/unconf/internal/config"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/spf13/cobra"
)

func newDBCmd() *cobra.Command {
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database utilities",
	}

	dbCmd.AddCommand(newDBStatusCmd())

	return dbCmd
}

func newDBStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.LoadConfig(cmd.Context(), config.LoadOptions{})
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			db, err := sqlite.NewConnectionManager(cmd.Context(), cfg.GetDBPath())
			if err != nil {
				return fmt.Errorf("failed to open database connection: %w", err)
			}
			defer func() {
				_ = db.Close()
			}()

			version, dirty, err := sqlite.MigrationStatus(db)
			if err != nil {
				return fmt.Errorf("failed to read migration status: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "db_path=%s\nmigration_version=%d\ndirty=%t\n", cfg.GetDBPath(), version, dirty)

			return nil
		},
	}
}
