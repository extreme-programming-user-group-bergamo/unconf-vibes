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

type roomsCommandClient struct {
	listClient    *client.Client
	bookingClient *client.AuthenticatedClient
}

type bookCommandClient struct {
	roomsClient *client.Client
	authClient  *client.AuthenticatedClient
}

type statusCommandClient struct {
	authClient *client.AuthenticatedClient
	apiClient  *client.Client
}

type attendeesCommandClient struct {
	authClient *client.AuthenticatedClient
}

func (c *roomsCommandClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	return c.listClient.ListRooms(ctx, slug)
}

func (c *roomsCommandClient) CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
	return c.bookingClient.CreateBooking(ctx, input)
}

func (c *bookCommandClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	return c.roomsClient.ListRooms(ctx, slug)
}

func (c *bookCommandClient) GetMe(ctx context.Context) (*client.UserResponse, error) {
	return c.authClient.GetMe(ctx)
}

func (c *bookCommandClient) CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
	return c.authClient.CreateBooking(ctx, input)
}

func (c *statusCommandClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	return c.authClient.ListBookings(ctx)
}

func (c *statusCommandClient) GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error) {
	return c.apiClient.GetConference(ctx, slug)
}

func (c *statusCommandClient) ListRoommateRequests(ctx context.Context) ([]client.RoommateRequestResponse, error) {
	return c.authClient.ListRoommateRequests(ctx)
}

func (c *attendeesCommandClient) ListAttendees(ctx context.Context, slug string) (*client.AttendeeListResponse, error) {
	return c.authClient.ListAttendees(ctx, slug)
}

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
	roomsClient := &roomsCommandClient{
		listClient:    apiClient,
		bookingClient: authClient,
	}
	bookClient := &bookCommandClient{
		roomsClient: apiClient,
		authClient:  authClient,
	}
	statusClient := &statusCommandClient{
		authClient: authClient,
		apiClient:  apiClient,
	}
	attendeesClient := &attendeesCommandClient{
		authClient: authClient,
	}

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
	rootCmd.AddCommand(newStatusCmd(statusClient, ctxManager))
	rootCmd.AddCommand(newAttendeesCmd(attendeesClient, ctxManager))
	rootCmd.AddCommand(newListCmd(apiClient))
	rootCmd.AddCommand(newInfoCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newCheckoutCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newRoomsCmd(roomsClient, ctxManager))
	rootCmd.AddCommand(newBookCmd(bookClient, ctxManager))
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
