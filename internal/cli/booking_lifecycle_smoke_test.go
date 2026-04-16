package cli

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/api"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/katurdays/unconf/internal/service"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type smokeContextStore struct {
	activeConference string
}

func (s *smokeContextStore) GetActiveConference() (string, error) {
	return s.activeConference, nil
}

func (s *smokeContextStore) SetActiveConference(slug string) error {
	s.activeConference = slug
	return nil
}

type smokeDeviceFlowProvider struct{}

func (p *smokeDeviceFlowProvider) StartDeviceFlow(_ context.Context) (*auth.DeviceAuthorization, error) {
	return &auth.DeviceAuthorization{
		DeviceCode:      "smoke-device-code",
		UserCode:        "SMOKE-1234",
		VerificationURI: "https://github.com/login/device",
		ExpiresIn:       120,
		Interval:        1,
	}, nil
}

func (p *smokeDeviceFlowProvider) ExchangeDeviceCode(_ context.Context, deviceCode string) (*auth.OAuthAccessToken, error) {
	if deviceCode != "smoke-device-code" {
		return nil, fmt.Errorf("failed to exchange device code: %w", service.ErrInvalidDeviceCode)
	}

	return &auth.OAuthAccessToken{AccessToken: "smoke-github-access-token"}, nil
}

func (p *smokeDeviceFlowProvider) FetchProfile(_ context.Context, accessToken string) (*auth.GitHubProfile, error) {
	if accessToken != "smoke-github-access-token" {
		return nil, fmt.Errorf("failed to fetch profile: invalid token")
	}

	return &auth.GitHubProfile{
		GitHubID:    "gh-smoke-booker",
		Email:       "smoke-booker@test.com",
		DisplayName: "Smoke Booker",
	}, nil
}

type bookingLifecycleSmokeHarness struct {
	server     *httptest.Server
	db         *sql.DB
	apiClient  *client.Client
	authClient *client.AuthenticatedClient
	store      *auth.MockTokenStore
}

func newBookingLifecycleSmokeHarness(t *testing.T) *bookingLifecycleSmokeHarness {
	t.Helper()

	db, err := sqlite.NewConnectionManager(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, sqlite.RunMigrations(db))

	userRepo := sqlite.NewUserRepository(db)
	refreshRepo := sqlite.NewRefreshSessionRepository(db)
	tokenService, err := auth.NewTokenService(bookTestSymmetricKey)
	require.NoError(t, err)

	authService, err := service.NewAuthService(
		&smokeDeviceFlowProvider{},
		tokenService,
		userRepo,
		refreshRepo,
		24*time.Hour,
		7*24*time.Hour,
	)
	require.NoError(t, err)

	conferenceRepo := sqlite.NewConferenceRepository(db)
	organizerRepo := sqlite.NewOrganizerRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	bookingRepo := sqlite.NewBookingRepository(db)
	requestRepo := sqlite.NewRoommateRequestRepository(db)

	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(service.NewUserService(userRepo))
	conferenceHandler := handlers.NewConferenceHandler(service.NewConferenceService(conferenceRepo, bookingRepo, organizerRepo))
	roomHandler := handlers.NewRoomHandler(service.NewRoomService(roomRepo, bookingRepo, conferenceRepo, userRepo))
	bookingHandler := handlers.NewBookingHandler(service.NewBookingService(bookingRepo, roomRepo, conferenceRepo, userRepo))
	requestHandler := handlers.NewRequestHandler(service.NewRequestService(requestRepo, bookingRepo, roomRepo, userRepo))

	router := api.NewRouter(authHandler, tokenService, userHandler, conferenceHandler, nil, nil, roomHandler, nil, bookingHandler, requestHandler)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	apiClient := client.NewClient(server.URL)
	store := auth.NewMockTokenStore()

	return &bookingLifecycleSmokeHarness{
		server:     server,
		db:         db,
		apiClient:  apiClient,
		authClient: client.NewAuthenticatedClient(apiClient, store),
		store:      store,
	}
}

func seedSmokeConferenceAndRoom(t *testing.T, db *sql.DB, slug string, roomNumber string, capacity int) (*models.Conference, *models.Room) {
	t.Helper()

	conferenceRepo := sqlite.NewConferenceRepository(db)
	conference, err := conferenceRepo.Create(context.Background(), &models.Conference{
		Slug:        slug,
		Name:        "SoCraTes 2026",
		Description: "Lifecycle smoke conference",
		Location:    "Bergamo",
		StartDate:   time.Now().Add(30 * 24 * time.Hour).UTC(),
		EndDate:     time.Now().Add(33 * 24 * time.Hour).UTC(),
		Capacity:    100,
	})
	require.NoError(t, err)

	roomRepo := sqlite.NewRoomRepository(db)
	room, err := roomRepo.Create(context.Background(), &models.Room{
		ConferenceID:  conference.ID,
		RoomNumber:    roomNumber,
		RoomType:      "double",
		PricePerNight: 150,
		Capacity:      capacity,
	})
	require.NoError(t, err)

	return conference, room
}

func runSmokeCommand(t *testing.T, cmd *cobra.Command, args []string, stdin string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(bytes.NewBufferString(stdin))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func TestBookingLifecycleSmoke_LoginCheckoutRoomsBookStatusCancel(t *testing.T) {
	harness := newBookingLifecycleSmokeHarness(t)
	_, _ = seedSmokeConferenceAndRoom(t, harness.db, "socrates-26", "101", 2)

	ctxStore := &smokeContextStore{}
	roomsClient := &roomsCommandClient{listClient: harness.apiClient, bookingClient: harness.authClient}
	bookClient := &bookCommandClient{roomsClient: harness.apiClient, authClient: harness.authClient}
	statusClient := &statusCommandClient{authClient: harness.authClient, apiClient: harness.apiClient}

	loginCmd := newLoginCmd(harness.apiClient, harness.store)
	loginOut, loginErrOut, err := runSmokeCommand(t, loginCmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, loginErrOut)
	assert.Contains(t, loginOut, "Welcome, Smoke Booker! You are now logged in.")

	checkoutCmd := newCheckoutCmd(harness.apiClient, ctxStore)
	checkoutOut, checkoutErrOut, err := runSmokeCommand(t, checkoutCmd, []string{"socrates-26"}, "")
	require.NoError(t, err)
	assert.Empty(t, checkoutErrOut)
	assert.Contains(t, checkoutOut, "Switched to socrates-26")

	roomsCmd := newRoomsCmdWithChecker(roomsClient, ctxStore, mockTerminalChecker{supports: false})
	roomsOut, roomsErrOut, err := runSmokeCommand(t, roomsCmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, roomsErrOut)
	assert.Contains(t, roomsOut, "Interactive room explorer unavailable")
	assert.Contains(t, roomsOut, "101")

	bookCmd := newBookCmd(bookClient, ctxStore)
	bookOut, bookErrOut, err := runSmokeCommand(t, bookCmd, []string{"101", "--private", "--notes", "late arrival", "--yes"}, "")
	require.NoError(t, err)
	assert.Empty(t, bookErrOut)
	assert.Contains(t, bookOut, "Booking created successfully.")
	assert.Contains(t, bookOut, "Status:     requested")

	roomsAfterBook, err := harness.apiClient.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, roomsAfterBook, 1)
	assert.Equal(t, 1, roomsAfterBook[0].SpotsTaken)
	assert.Equal(t, 1, roomsAfterBook[0].SpotsAvailable)

	statusCmd := newStatusCmd(statusClient, ctxStore)
	statusOut, statusErrOut, err := runSmokeCommand(t, statusCmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, statusErrOut)
	assert.Contains(t, statusOut, `Booking status for conference "socrates-26":`)
	assert.Contains(t, statusOut, "Status:     requested")
	assert.Contains(t, statusOut, "Room:       Room 101")

	cancelCmd := newCancelCmd(bookClient, ctxStore)
	cancelOut, cancelErrOut, err := runSmokeCommand(t, cancelCmd, []string{"--yes"}, "")
	require.NoError(t, err)
	assert.Empty(t, cancelErrOut)
	assert.Contains(t, cancelOut, "Booking cancelled successfully.")

	statusAfterCancelOut, statusAfterCancelErrOut, err := runSmokeCommand(t, statusCmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, statusAfterCancelErrOut)
	assert.Contains(t, statusAfterCancelOut, `No booking found for active conference "socrates-26"`)

	roomsAfterCancel, err := harness.apiClient.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, roomsAfterCancel, 1)
	assert.Equal(t, 0, roomsAfterCancel[0].SpotsTaken)
	assert.Equal(t, 2, roomsAfterCancel[0].SpotsAvailable)
}

func TestBookingLifecycleSmoke_BookReportsRoomFullAndKeepsStatusEmpty(t *testing.T) {
	harness := newBookingLifecycleSmokeHarness(t)
	conference, room := seedSmokeConferenceAndRoom(t, harness.db, "socrates-26", "101", 1)

	userRepo := sqlite.NewUserRepository(harness.db)
	occupant, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:       "gh-existing-occupant",
		Email:          "occupant@test.com",
		DisplayName:    "Existing Occupant",
		PrivacySetting: "public",
	})
	require.NoError(t, err)

	bookingRepo := sqlite.NewBookingRepository(harness.db)
	_, err = bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         room.ID,
		UserID:         occupant.ID,
		ConferenceID:   conference.ID,
		Status:         models.BookingStatusConfirmed,
		PrivacySetting: "public",
	})
	require.NoError(t, err)

	ctxStore := &smokeContextStore{}
	bookClient := &bookCommandClient{roomsClient: harness.apiClient, authClient: harness.authClient}
	statusClient := &statusCommandClient{authClient: harness.authClient, apiClient: harness.apiClient}

	loginCmd := newLoginCmd(harness.apiClient, harness.store)
	_, _, err = runSmokeCommand(t, loginCmd, []string{}, "")
	require.NoError(t, err)

	checkoutCmd := newCheckoutCmd(harness.apiClient, ctxStore)
	_, _, err = runSmokeCommand(t, checkoutCmd, []string{"socrates-26"}, "")
	require.NoError(t, err)

	bookCmd := newBookCmd(bookClient, ctxStore)
	bookOut, bookErrOut, err := runSmokeCommand(t, bookCmd, []string{"101", "--private", "--yes"}, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create booking for room 101 in conference socrates-26")
	assert.Contains(t, bookErrOut, "Room 101 is full in conference socrates-26")
	assert.NotContains(t, bookOut, "Booking created successfully")

	statusCmd := newStatusCmd(statusClient, ctxStore)
	statusOut, statusErrOut, err := runSmokeCommand(t, statusCmd, []string{}, "")
	require.NoError(t, err)
	assert.Empty(t, statusErrOut)
	assert.Contains(t, statusOut, `No booking found for active conference "socrates-26"`)

	roomsAfterAttempt, err := harness.apiClient.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, roomsAfterAttempt, 1)
	assert.Equal(t, 1, roomsAfterAttempt[0].SpotsTaken)
	assert.Equal(t, 0, roomsAfterAttempt[0].SpotsAvailable)
}
