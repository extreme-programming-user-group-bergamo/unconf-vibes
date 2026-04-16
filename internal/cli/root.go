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

type dashboardCommandClient struct {
	authClient *client.AuthenticatedClient
}

type inviteCommandClient struct {
	authClient *client.AuthenticatedClient
}

type requestsCommandClient struct {
	authClient *client.AuthenticatedClient
}

type createConferenceCommandClient struct {
	authClient *client.AuthenticatedClient
}

type editConferenceCommandClient struct {
	authClient *client.AuthenticatedClient
	apiClient  *client.Client
}

type exportCommandClient struct {
	authClient *client.AuthenticatedClient
}

type sessionsCommandClient struct {
	authClient *client.AuthenticatedClient
}

func (c *roomsCommandClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	return c.listClient.ListRooms(ctx, slug)
}

func (c *roomsCommandClient) CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
	return c.bookingClient.CreateBooking(ctx, input)
}

func (c *roomsCommandClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	return c.bookingClient.ListBookings(ctx)
}

func (c *roomsCommandClient) CreateRoommateRequest(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
	return c.bookingClient.CreateRoommateRequest(ctx, input)
}

func (c *roomsCommandClient) CreateRoom(ctx context.Context, slug string, input client.ManageRoomRequest) (*client.RoomResponse, error) {
	return c.bookingClient.CreateRoom(ctx, slug, input)
}

func (c *roomsCommandClient) UpdateRoom(ctx context.Context, slug string, roomNumber string, input client.ManageRoomRequest) (*client.RoomResponse, error) {
	return c.bookingClient.UpdateRoom(ctx, slug, roomNumber, input)
}

func (c *roomsCommandClient) DeleteRoom(ctx context.Context, slug string, roomNumber string) error {
	return c.bookingClient.DeleteRoom(ctx, slug, roomNumber)
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

func (c *bookCommandClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	return c.authClient.ListBookings(ctx)
}

func (c *bookCommandClient) CancelBooking(ctx context.Context, bookingID int64) (*client.BookingResponse, error) {
	return c.authClient.CancelBooking(ctx, bookingID)
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

func (c *dashboardCommandClient) GetOrganizerDashboard(ctx context.Context, slug string, query client.DashboardQuery) (*client.OrganizerDashboardResponse, error) {
	return c.authClient.GetOrganizerDashboard(ctx, slug, query)
}

func (c *inviteCommandClient) ListBookings(ctx context.Context) ([]client.BookingResponse, error) {
	return c.authClient.ListBookings(ctx)
}

func (c *inviteCommandClient) CreateRoommateRequest(ctx context.Context, input client.CreateRoommateRequestRequest) (*client.RoommateRequestResponse, error) {
	return c.authClient.CreateRoommateRequest(ctx, input)
}

func (c *requestsCommandClient) ListRoommateRequests(ctx context.Context) ([]client.RoommateRequestResponse, error) {
	return c.authClient.ListRoommateRequests(ctx)
}

func (c *requestsCommandClient) AcceptRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
	return c.authClient.AcceptRoommateRequest(ctx, requestID)
}

func (c *requestsCommandClient) DeclineRoommateRequest(ctx context.Context, requestID int64) (*client.RoommateRequestResponse, error) {
	return c.authClient.DeclineRoommateRequest(ctx, requestID)
}

func (c *createConferenceCommandClient) CreateConference(ctx context.Context, input client.CreateConferenceRequest) (*client.ConferenceResponse, error) {
	return c.authClient.CreateConference(ctx, input)
}

func (c *editConferenceCommandClient) GetConference(ctx context.Context, slug string) (*client.ConferenceResponse, error) {
	return c.apiClient.GetConference(ctx, slug)
}

func (c *editConferenceCommandClient) UpdateConference(ctx context.Context, slug string, input client.UpdateConferenceRequest) (*client.ConferenceResponse, error) {
	return c.authClient.UpdateConference(ctx, slug, input)
}

func (c *exportCommandClient) ExportConferenceBookingsCSV(ctx context.Context, slug string, includeCancelled bool) ([]byte, error) {
	return c.authClient.ExportConferenceBookingsCSV(ctx, slug, includeCancelled)
}

func (c *sessionsCommandClient) ListSessions(ctx context.Context) ([]client.SessionResponse, error) {
	return c.authClient.ListSessions(ctx)
}

func (c *sessionsCommandClient) RevokeSession(ctx context.Context, sessionID int64) error {
	return c.authClient.RevokeSession(ctx, sessionID)
}

func (c *sessionsCommandClient) RevokeOtherSessions(ctx context.Context) (int64, error) {
	return c.authClient.RevokeOtherSessions(ctx)
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

			slog.Info("UNCONF CLI starting",
				"version", Version,
				"config_file", cfg.GetConfigFile(),
				"api_endpoint", cfg.GetAPIEndpoint(),
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
	dashboardClient := &dashboardCommandClient{
		authClient: authClient,
	}
	inviteClient := &inviteCommandClient{
		authClient: authClient,
	}
	requestsClient := &requestsCommandClient{
		authClient: authClient,
	}
	createConferenceClient := &createConferenceCommandClient{
		authClient: authClient,
	}
	editConferenceClient := &editConferenceCommandClient{
		authClient: authClient,
		apiClient:  apiClient,
	}
	exportClient := &exportCommandClient{
		authClient: authClient,
	}
	sessionsClient := &sessionsCommandClient{
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

	rootCmd.AddCommand(newLoginCmd(apiClient, store))
	rootCmd.AddCommand(newLogoutCmd(apiClient, store))
	rootCmd.AddCommand(newStatusCmd(statusClient, ctxManager))
	rootCmd.AddCommand(newAttendeesCmd(attendeesClient, ctxManager))
	rootCmd.AddCommand(newDashboardCmd(dashboardClient, ctxManager))
	rootCmd.AddCommand(newInviteCmd(inviteClient, ctxManager))
	rootCmd.AddCommand(newRequestsCmd(requestsClient))
	rootCmd.AddCommand(newListCmd(apiClient))
	rootCmd.AddCommand(newInfoCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newCheckoutCmd(apiClient, ctxManager))
	rootCmd.AddCommand(newRoomsCmd(roomsClient, ctxManager))
	rootCmd.AddCommand(newBookCmd(bookClient, ctxManager))
	rootCmd.AddCommand(newCancelCmd(bookClient, ctxManager))
	rootCmd.AddCommand(newConfigCmd(authClient))
	rootCmd.AddCommand(newCreateCmd(createConferenceClient))
	rootCmd.AddCommand(newEditCmd(editConferenceClient, ctxManager))
	rootCmd.AddCommand(newExportCmd(exportClient, ctxManager))
	rootCmd.AddCommand(newSessionsCmd(sessionsClient))

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
