package cli

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
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

const bookTestSymmetricKey = "0123456789abcdef0123456789abcdef"

type mockBookClient struct {
	listRoomsFn     func(ctx context.Context, slug string) ([]client.RoomResponse, error)
	getMeFn         func(ctx context.Context) (*client.UserResponse, error)
	createBookingFn func(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error)
}

func (m *mockBookClient) ListRooms(ctx context.Context, slug string) ([]client.RoomResponse, error) {
	if m.listRoomsFn != nil {
		return m.listRoomsFn(ctx, slug)
	}
	return []client.RoomResponse{}, nil
}

func (m *mockBookClient) GetMe(ctx context.Context) (*client.UserResponse, error) {
	if m.getMeFn != nil {
		return m.getMeFn(ctx)
	}
	return &client.UserResponse{PrivacySetting: "public"}, nil
}

func (m *mockBookClient) CreateBooking(ctx context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
	if m.createBookingFn != nil {
		return m.createBookingFn(ctx, input)
	}
	return &client.BookingResponse{ID: 1, Status: "confirmed", PrivacySetting: input.PrivacySetting, Notes: input.Notes}, nil
}

type mockBookContextStore struct {
	getActiveConferenceFn func() (string, error)
}

func (m *mockBookContextStore) GetActiveConference() (string, error) {
	if m.getActiveConferenceFn != nil {
		return m.getActiveConferenceFn()
	}
	return "", nil
}

type bookIntegrationHarness struct {
	server       *httptest.Server
	db           *sql.DB
	tokenService *auth.TokenService
}

func newBookIntegrationHarness(t *testing.T) *bookIntegrationHarness {
	t.Helper()

	db, err := sqlite.NewConnectionManager(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, sqlite.RunMigrations(db))

	tokenService, err := auth.NewTokenService(bookTestSymmetricKey)
	require.NoError(t, err)

	userRepo := sqlite.NewUserRepository(db)
	conferenceRepo := sqlite.NewConferenceRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	bookingRepo := sqlite.NewBookingRepository(db)

	userHandler := handlers.NewUserHandler(service.NewUserService(userRepo))
	roomHandler := handlers.NewRoomHandler(service.NewRoomService(roomRepo, bookingRepo, conferenceRepo, userRepo))
	bookingHandler := handlers.NewBookingHandler(service.NewBookingService(bookingRepo, roomRepo, conferenceRepo, userRepo))

	router := api.NewRouter(nil, tokenService, userHandler, nil, nil, nil, roomHandler, nil, bookingHandler, nil)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &bookIntegrationHarness{server: server, db: db, tokenService: tokenService}
}

func createBookIntegrationUser(t *testing.T, db *sql.DB, githubID, email, displayName, privacy string) *models.User {
	t.Helper()

	userRepo := sqlite.NewUserRepository(db)
	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:       githubID,
		Email:          email,
		DisplayName:    displayName,
		PrivacySetting: privacy,
	})
	require.NoError(t, err)

	return user
}

func createBookIntegrationConference(t *testing.T, db *sql.DB, slug string) *models.Conference {
	t.Helper()

	conferenceRepo := sqlite.NewConferenceRepository(db)
	conference, err := conferenceRepo.Create(context.Background(), &models.Conference{
		Slug:        slug,
		Name:        "SoCraTes 2026",
		Description: "Integration test conference",
		Location:    "Bergamo",
		StartDate:   time.Now().Add(30 * 24 * time.Hour).UTC(),
		EndDate:     time.Now().Add(33 * 24 * time.Hour).UTC(),
		Capacity:    120,
	})
	require.NoError(t, err)

	return conference
}

func createBookIntegrationRoom(t *testing.T, db *sql.DB, conferenceID int64, roomNumber string) *models.Room {
	t.Helper()

	roomRepo := sqlite.NewRoomRepository(db)
	room, err := roomRepo.Create(context.Background(), &models.Room{
		ConferenceID:  conferenceID,
		RoomNumber:    roomNumber,
		RoomType:      "double",
		PricePerNight: 120,
		Capacity:      2,
	})
	require.NoError(t, err)

	return room
}

func createBookIntegrationSession(t *testing.T, db *sql.DB, userID int64, tokenService *auth.TokenService) string {
	t.Helper()

	refreshRepo := sqlite.NewRefreshSessionRepository(db)
	now := time.Now().UTC()

	session, err := refreshRepo.Create(context.Background(), &models.RefreshSession{
		UserID:        userID,
		TokenHash:     "book-test-hash",
		ExpiresAt:     now.Add(7 * 24 * time.Hour),
		IssuedAt:      now,
		LastAccessJTI: "book-test-jti",
	})
	require.NoError(t, err)

	accessToken, err := tokenService.IssueAccessToken(context.Background(), auth.AccessTokenInput{
		UserID:     userID,
		SessionID:  session.ID,
		Issuer:     "unconf-api",
		Audience:   "unconf-cli",
		NotBefore:  now,
		IssuedAt:   now,
		JTI:        "book-test-access-jti",
		ExpiryTime: now.Add(1 * time.Hour),
	})
	require.NoError(t, err)

	return accessToken
}

func executeBookCmd(t *testing.T, cmd *cobra.Command, args []string, stdin string) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(strings.NewReader(stdin))

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func executeBookCmdWithInput(t *testing.T, cmd *cobra.Command, args []string, input io.Reader) (string, string, error) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	cmd.SetIn(input)

	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

type errReader struct {
	err error
}

func (r *errReader) Read(_ []byte) (int, error) {
	return 0, r.err
}

func TestBookCmd_CreatesBookingWithProfilePrivacyAndNotes(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{
		getActiveConferenceFn: func() (string, error) {
			return "socrates-26", nil
		},
	}

	var capturedInput client.CreateBookingRequest
	bookClient.listRoomsFn = func(_ context.Context, slug string) ([]client.RoomResponse, error) {
		require.Equal(t, "socrates-26", slug)
		return []client.RoomResponse{{ID: 42, ConferenceID: 7, RoomNumber: "101"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "private"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		capturedInput = input
		return &client.BookingResponse{
			ID:             9,
			Status:         "confirmed",
			PrivacySetting: input.PrivacySetting,
			Notes:          input.Notes,
		}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"101", "--notes", "vegetarian", "--yes"}, "")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Equal(t, int64(42), capturedInput.RoomID)
	assert.Equal(t, int64(7), capturedInput.ConferenceID)
	assert.Equal(t, "private", capturedInput.PrivacySetting)
	assert.Equal(t, "vegetarian", capturedInput.Notes)
	assert.Contains(t, stdout, "Booking created successfully")
	assert.Contains(t, stdout, "Room:       101")
}

func TestBookCmd_CreatesBookingAgainstRunningServer(t *testing.T) {
	harness := newBookIntegrationHarness(t)
	conference := createBookIntegrationConference(t, harness.db, "socrates-26")
	room := createBookIntegrationRoom(t, harness.db, conference.ID, "101")
	user := createBookIntegrationUser(t, harness.db, "gh-book-cli", "book-cli@test.com", "CLI Booker", "private")
	accessToken := createBookIntegrationSession(t, harness.db, user.ID, harness.tokenService)

	store := auth.NewMockTokenStore()
	store.SetTokens(accessToken, "valid-refresh-token")
	apiClient := client.NewClient(harness.server.URL)
	bookClient := &bookCommandClient{
		roomsClient: apiClient,
		authClient:  client.NewAuthenticatedClient(apiClient, store),
	}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"101", "--yes"}, "")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Booking created successfully")
	assert.Contains(t, stdout, "Room:       101")
	assert.Contains(t, stdout, "Status:     requested")

	bookingRepo := sqlite.NewBookingRepository(harness.db)
	createdBooking, err := bookingRepo.GetActiveByUserAndConference(context.Background(), user.ID, conference.ID)
	require.NoError(t, err)
	assert.Equal(t, room.ID, createdBooking.RoomID)
	assert.Equal(t, "private", createdBooking.PrivacySetting)
	assert.Equal(t, models.BookingStatusRequested, createdBooking.Status)

	rooms, err := apiClient.ListRooms(context.Background(), "socrates-26")
	require.NoError(t, err)
	require.Len(t, rooms, 1)
	assert.Equal(t, 1, rooms[0].SpotsTaken)
	assert.Equal(t, 1, rooms[0].SpotsAvailable)

	bookings, err := bookClient.ListBookings(context.Background())
	require.NoError(t, err)
	require.Len(t, bookings, 1)
	assert.Equal(t, createdBooking.ID, bookings[0].ID)
	assert.Equal(t, "requested", bookings[0].Status)
	assert.Equal(t, "socrates-26", bookings[0].ConferenceSlug)
}

func TestBookCmd_PrivateFlagOverridesProfileDefault(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 10, ConferenceID: 8, RoomNumber: "202"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		assert.Equal(t, "private", input.PrivacySetting)
		return &client.BookingResponse{ID: 12, Status: "confirmed", PrivacySetting: input.PrivacySetting}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	_, _, err := executeBookCmd(t, cmd, []string{"202", "--private", "--yes"}, "")
	require.NoError(t, err)
}

func TestBookCmd_ConfirmationPromptCancelSkipsCreate(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 1, ConferenceID: 2, RoomNumber: "303"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, fmt.Errorf("should not be called")
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"303"}, "n\n")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Book room 303")
	assert.Contains(t, stdout, "Booking canceled")
}

func TestBookCmd_CreateBookingErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		roomNumber     string
		createErr      error
		expectedStderr string
		expectedErrIs  error
	}{
		{
			name:           "room full",
			roomNumber:     "404",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrRoomFull),
			expectedStderr: "Room 404 is full",
			expectedErrIs:  client.ErrRoomFull,
		},
		{
			name:           "already booked",
			roomNumber:     "505",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrAlreadyBooked),
			expectedStderr: "already have a booking",
			expectedErrIs:  client.ErrAlreadyBooked,
		},
		{
			name:           "create path room not found",
			roomNumber:     "606",
			createErr:      fmt.Errorf("create booking failed: %w", client.ErrRoomNotFound),
			expectedStderr: "Room 606 was not found in conference socrates-26",
			expectedErrIs:  client.ErrRoomNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bookClient := &mockBookClient{}
			ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

			bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
				return []client.RoomResponse{{ID: 1, ConferenceID: 2, RoomNumber: tt.roomNumber}}, nil
			}
			bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
				return &client.UserResponse{PrivacySetting: "public"}, nil
			}
			bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
				return nil, tt.createErr
			}

			cmd := newBookCmd(bookClient, ctxStore)
			_, stderr, err := executeBookCmd(t, cmd, []string{tt.roomNumber, "--yes"}, "")

			require.Error(t, err)
			assert.Contains(t, stderr, tt.expectedStderr)
			if errors.Is(tt.createErr, client.ErrRoomNotFound) {
				assert.Contains(t, stderr, "Refresh room list and try again")
			}
			assert.Contains(t, err.Error(), "failed to create booking")
			assert.ErrorIs(t, err, tt.expectedErrIs)
		})
	}
}

func TestBookCmd_ConfirmationPromptAcceptsEOFWithInput(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 11, ConferenceID: 22, RoomNumber: "707"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		return &client.BookingResponse{ID: 1, Status: "confirmed", PrivacySetting: input.PrivacySetting}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	stdout, stderr, err := executeBookCmd(t, cmd, []string{"707"}, "yes")

	require.NoError(t, err)
	assert.Empty(t, stderr)
	assert.Contains(t, stdout, "Booking created successfully")
}

func TestBookCmd_ConfirmationPromptReadErrorReturnsFailure(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 13, ConferenceID: 24, RoomNumber: "808"}}, nil
	}
	bookClient.getMeFn = func(_ context.Context) (*client.UserResponse, error) {
		return &client.UserResponse{PrivacySetting: "public"}, nil
	}
	bookClient.createBookingFn = func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, fmt.Errorf("should not be called")
	}

	readErr := errors.New("stdin exploded")
	cmd := newBookCmd(bookClient, ctxStore)
	_, stderr, err := executeBookCmdWithInput(t, cmd, []string{"808"}, &errReader{err: readErr})

	require.Error(t, err)
	assert.Contains(t, stderr, "failed to confirm booking")
	assert.Contains(t, err.Error(), "failed to confirm booking")
	assert.ErrorIs(t, err, readErr)
}

func TestBookCmd_RoomNotFoundByArgument(t *testing.T) {
	bookClient := &mockBookClient{}
	ctxStore := &mockBookContextStore{getActiveConferenceFn: func() (string, error) { return "socrates-26", nil }}

	bookClient.listRoomsFn = func(_ context.Context, _ string) ([]client.RoomResponse, error) {
		return []client.RoomResponse{{ID: 7, ConferenceID: 8, RoomNumber: "101"}}, nil
	}

	cmd := newBookCmd(bookClient, ctxStore)
	_, _, err := executeBookCmd(t, cmd, []string{"999", "--yes"}, "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "room \"999\" not found")
	assert.ErrorIs(t, err, client.ErrRoomNotFound)
}
