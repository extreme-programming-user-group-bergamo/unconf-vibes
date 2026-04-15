package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/katurdays/unconf/internal/service"
)

const testSymmetricKey = "0123456789abcdef0123456789abcdef"

func setupIntegrationRouter(t *testing.T) (*httptest.Server, *auth.TokenService, *sql.DB) {
	t.Helper()

	db, err := sqlite.NewConnectionManager(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	require.NoError(t, sqlite.RunMigrations(db))

	userRepo := sqlite.NewUserRepository(db)
	refreshRepo := sqlite.NewRefreshSessionRepository(db)

	tokenService, err := auth.NewTokenService(testSymmetricKey)
	require.NoError(t, err)

	authService, err := service.NewAuthService(
		&stubDeviceFlowProvider{},
		tokenService,
		userRepo,
		refreshRepo,
		24*time.Hour,
		7*24*time.Hour,
	)
	require.NoError(t, err)

	userService := service.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)

	conferenceRepo := sqlite.NewConferenceRepository(db)
	organizerRepo := sqlite.NewOrganizerRepository(db)
	bookingRepo := sqlite.NewBookingRepository(db)
	conferenceService := service.NewConferenceService(conferenceRepo, bookingRepo, organizerRepo)
	conferenceHandler := handlers.NewConferenceHandler(conferenceService)
	organizerService := service.NewOrganizerService(conferenceRepo, organizerRepo, userRepo)
	organizerHandler := handlers.NewOrganizerHandler(organizerService)

	roomRepo := sqlite.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepo, bookingRepo, conferenceRepo, userRepo)
	roomHandler := handlers.NewRoomHandler(roomService)
	attendeeService := service.NewAttendeeService(conferenceRepo, organizerRepo, bookingRepo, roomRepo, userRepo)
	attendeeHandler := handlers.NewAttendeeHandler(attendeeService)
	requestRepo := sqlite.NewRoommateRequestRepository(db)
	bookingService := service.NewBookingService(bookingRepo, roomRepo, conferenceRepo, userRepo)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	requestService := service.NewRequestService(requestRepo, bookingRepo, roomRepo, userRepo)
	requestHandler := handlers.NewRequestHandler(requestService)

	router := NewRouter(authHandler, tokenService, userHandler, conferenceHandler, organizerHandler, organizerService, roomHandler, attendeeHandler, bookingHandler, requestHandler)
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return srv, tokenService, db
}

func createTestUser(t *testing.T, db *sql.DB) *models.User {
	t.Helper()

	userRepo := sqlite.NewUserRepository(db)
	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:       "test-github-123",
		Email:          "test@example.com",
		DisplayName:    "Test User",
		PrivacySetting: "public",
	})
	require.NoError(t, err)

	return user
}

func createTestSession(t *testing.T, db *sql.DB, userID int64, tokenService *auth.TokenService) (string, int64) {
	t.Helper()

	refreshRepo := sqlite.NewRefreshSessionRepository(db)
	now := time.Now().UTC()

	session, err := refreshRepo.Create(context.Background(), &models.RefreshSession{
		UserID:        userID,
		TokenHash:     "test-hash",
		ExpiresAt:     now.Add(7 * 24 * time.Hour),
		IssuedAt:      now,
		LastAccessJTI: "test-jti",
	})
	require.NoError(t, err)

	accessToken, err := tokenService.IssueAccessToken(context.Background(), auth.AccessTokenInput{
		UserID:     userID,
		SessionID:  session.ID,
		Issuer:     "unconf-api",
		Audience:   "unconf-cli",
		NotBefore:  now,
		IssuedAt:   now,
		JTI:        "integration-test-jti",
		ExpiryTime: now.Add(1 * time.Hour),
	})
	require.NoError(t, err)

	return accessToken, session.ID
}

type stubDeviceFlowProvider struct{}

func (s *stubDeviceFlowProvider) StartDeviceFlow(_ context.Context) (*auth.DeviceAuthorization, error) {
	return &auth.DeviceAuthorization{}, nil
}

func (s *stubDeviceFlowProvider) ExchangeDeviceCode(_ context.Context, _ string) (*auth.OAuthAccessToken, error) {
	return nil, nil
}

func (s *stubDeviceFlowProvider) FetchProfile(_ context.Context, _ string) (*auth.GitHubProfile, error) {
	return nil, nil
}

func TestProtectedEndpoint_NoAuthHeader_Returns401(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/users/me")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestProtectedEndpoint_InvalidToken_Returns401(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-garbage-token")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetUsersMe_ValidToken_Returns200(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "Test User", body["display_name"])
	assert.Equal(t, "test@example.com", body["email"])
}

func TestPutUsersMe_ValidToken_Returns200(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"display_name":"Updated Name","privacy_setting":"private"}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/users/me", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "Updated Name", result["display_name"])
	assert.Equal(t, "private", result["privacy_setting"])
}

func TestPostAuthRevoke_ValidToken_Returns204(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/auth/revoke", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestPostAuthRevoke_ThenRevoke_SessionAlreadyRevoked(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	// First revoke should succeed
	req1, _ := http.NewRequest(http.MethodPost, srv.URL+"/auth/revoke", nil)
	req1.Header.Set("Authorization", "Bearer "+accessToken)

	resp1, err := http.DefaultClient.Do(req1)
	require.NoError(t, err)
	defer func() { _ = resp1.Body.Close() }()
	assert.Equal(t, http.StatusNoContent, resp1.StatusCode)

	// Second revoke with same token should fail
	req2, _ := http.NewRequest(http.MethodPost, srv.URL+"/auth/revoke", nil)
	req2.Header.Set("Authorization", "Bearer "+accessToken)

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
}

func TestHealthEndpoint_NoAuth_Returns200(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/health")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func createTestSessionWithRefreshToken(t *testing.T, db *sql.DB, userID int64, rawRefreshToken string, expiresAt time.Time) int64 {
	t.Helper()

	refreshRepo := sqlite.NewRefreshSessionRepository(db)
	now := time.Now().UTC()

	session, err := refreshRepo.Create(context.Background(), &models.RefreshSession{
		UserID:        userID,
		TokenHash:     auth.HashRefreshToken(rawRefreshToken),
		ExpiresAt:     expiresAt,
		IssuedAt:      now,
		LastAccessJTI: "test-refresh-jti",
	})
	require.NoError(t, err)

	return session.ID
}

func TestPostAuthRefresh_ValidToken_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	rawRefreshToken := "valid-raw-refresh-token-for-test"
	createTestSessionWithRefreshToken(t, db, user.ID, rawRefreshToken, time.Now().UTC().Add(7*24*time.Hour))

	body := `{"refresh_token":"` + rawRefreshToken + `"}`
	resp, err := http.Post(srv.URL+"/auth/refresh", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["access_token"])
	assert.NotEmpty(t, result["refresh_token"])
	assert.Equal(t, "Bearer", result["token_type"])
	assert.NotNil(t, result["user"])

	// Verify old refresh token no longer works (rotation invalidates it)
	resp2, err := http.Post(srv.URL+"/auth/refresh", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)

	var errResp map[string]interface{}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&errResp))
	errObj := errResp["error"].(map[string]interface{})
	assert.Equal(t, "revoked_refresh_token", errObj["code"])
}

func TestPostAuthRefresh_InvalidToken_Returns401(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	body := `{"refresh_token":"completely-invalid-token"}`
	resp, err := http.Post(srv.URL+"/auth/refresh", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errResp map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	errObj := errResp["error"].(map[string]interface{})
	assert.Equal(t, "invalid_refresh_token", errObj["code"])
}

func TestPostAuthRefresh_ExpiredToken_Returns401(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	rawRefreshToken := "expired-raw-refresh-token-for-test"
	// Create session with expiry in the past
	createTestSessionWithRefreshToken(t, db, user.ID, rawRefreshToken, time.Now().UTC().Add(-1*time.Hour))

	body := `{"refresh_token":"` + rawRefreshToken + `"}`
	resp, err := http.Post(srv.URL+"/auth/refresh", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errResp map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	errObj := errResp["error"].(map[string]interface{})
	assert.Equal(t, "expired_refresh_token", errObj["code"])
}

func TestPostAuthRefresh_ThenUseNewAccessToken_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	user := createTestUser(t, db)
	rawRefreshToken := "roundtrip-raw-refresh-token"
	createTestSessionWithRefreshToken(t, db, user.ID, rawRefreshToken, time.Now().UTC().Add(7*24*time.Hour))

	// Step 1: Refresh to get new access token
	body := `{"refresh_token":"` + rawRefreshToken + `"}`
	resp, err := http.Post(srv.URL+"/auth/refresh", "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	newAccessToken := result["access_token"].(string)
	require.NotEmpty(t, newAccessToken)

	// Step 2: Use new access token on protected endpoint
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+newAccessToken)

	meResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = meResp.Body.Close() }()

	assert.Equal(t, http.StatusOK, meResp.StatusCode)

	var meBody map[string]interface{}
	require.NoError(t, json.NewDecoder(meResp.Body).Decode(&meBody))
	assert.Equal(t, "Test User", meBody["display_name"])
}

func createTestConference(t *testing.T, db *sql.DB, slug, name string, startDate, endDate time.Time) {
	t.Helper()

	confRepo := sqlite.NewConferenceRepository(db)
	_, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        slug,
		Name:        name,
		Description: "Test conference description",
		Location:    "Test City",
		StartDate:   startDate,
		EndDate:     endDate,
		Capacity:    100,
	})
	require.NoError(t, err)
}

func createTestOrganizer(t *testing.T, db *sql.DB, conferenceID int64, userID int64, role models.ConferenceOrganizerRole) {
	t.Helper()

	repo := sqlite.NewOrganizerRepository(db)
	_, err := repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conferenceID,
		UserID:       userID,
		Role:         role,
	})
	require.NoError(t, err)
}

func TestGetConferences_Empty_Returns200(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/conferences")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Empty(t, body)
}

func TestGetConferences_WithData_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-a", "Conference A",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	createTestConference(t, db, "conf-b", "Conference B",
		time.Now().Add(-30*24*time.Hour), time.Now().Add(-27*24*time.Hour))

	resp, err := http.Get(srv.URL + "/conferences")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 2)

	// Ordered by start_date DESC — upcoming first
	assert.Equal(t, "conf-a", body[0]["slug"])
	assert.Equal(t, "upcoming", body[0]["status"])
	assert.Equal(t, float64(0), body[0]["attendee_count"])
	assert.Equal(t, "conf-b", body[1]["slug"])
	assert.Equal(t, "past", body[1]["status"])
}

func TestGetConferenceBySlug_Found_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "socrates-26", "SoCraTes 2026",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	resp, err := http.Get(srv.URL + "/conferences/socrates-26")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "socrates-26", body["slug"])
	assert.Equal(t, "SoCraTes 2026", body["name"])
	assert.Equal(t, "upcoming", body["status"])
	assert.Equal(t, float64(0), body["attendee_count"])
}

func TestGetConferenceBySlug_NotFound_Returns404(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/conferences/nonexistent")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	errObj := body["error"].(map[string]interface{})
	assert.Equal(t, "not_found", errObj["code"])
}

func TestGetConferences_NoAuthRequired(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	// No Authorization header — should still return 200
	resp, err := http.Get(srv.URL + "/conferences")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetConferences_StatusDerived(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	// Past conference
	createTestConference(t, db, "past-conf", "Past Conf",
		time.Now().Add(-60*24*time.Hour), time.Now().Add(-57*24*time.Hour))
	// Active conference
	createTestConference(t, db, "active-conf", "Active Conf",
		time.Now().Add(-1*24*time.Hour), time.Now().Add(2*24*time.Hour))
	// Upcoming conference
	createTestConference(t, db, "upcoming-conf", "Upcoming Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	resp, err := http.Get(srv.URL + "/conferences")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 3)

	// Collect statuses by slug
	statuses := make(map[string]string)
	for _, c := range body {
		statuses[c["slug"].(string)] = c["status"].(string)
	}

	assert.Equal(t, "past", statuses["past-conf"])
	assert.Equal(t, "active", statuses["active-conf"])
	assert.Equal(t, "upcoming", statuses["upcoming-conf"])
}

// --- Room integration test helpers ---

func createTestRoom(t *testing.T, db *sql.DB, conferenceID int64, roomNumber, roomType string, pricePerNight float64, capacity int) *models.Room {
	t.Helper()

	roomRepo := sqlite.NewRoomRepository(db)
	room, err := roomRepo.Create(context.Background(), &models.Room{
		ConferenceID:  conferenceID,
		RoomNumber:    roomNumber,
		RoomType:      roomType,
		PricePerNight: pricePerNight,
		Capacity:      capacity,
	})
	require.NoError(t, err)

	return room
}

func createTestBooking(t *testing.T, db *sql.DB, roomID, userID, conferenceID int64, status models.BookingStatus, privacySetting string) *models.Booking {
	t.Helper()

	bookingRepo := sqlite.NewBookingRepository(db)
	booking, err := bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         roomID,
		UserID:         userID,
		ConferenceID:   conferenceID,
		Status:         status,
		PrivacySetting: privacySetting,
	})
	require.NoError(t, err)

	return booking
}

func createTestUserWithName(t *testing.T, db *sql.DB, githubID, email, displayName, privacy string) *models.User {
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

// --- Room integration tests ---

func TestGetRooms_Empty_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-rooms", "Rooms Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	resp, err := http.Get(srv.URL + "/conferences/conf-rooms/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Empty(t, body)
}

func TestGetRooms_WithRooms_Returns200(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-rooms2", "Rooms Conf 2",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	// Need conference ID — fetch it
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-rooms2")
	require.NoError(t, err)

	createTestRoom(t, db, conf.ID, "101", "double", 120.0, 2)
	createTestRoom(t, db, conf.ID, "102", "single", 80.0, 1)

	resp, err := http.Get(srv.URL + "/conferences/conf-rooms2/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 2)

	assert.Equal(t, "101", body[0]["room_number"])
	assert.Equal(t, float64(2), body[0]["capacity"])
	assert.Equal(t, float64(0), body[0]["spots_taken"])
	assert.Equal(t, float64(2), body[0]["spots_available"])
}

func TestGetRooms_AvailabilityComputed(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-avail", "Avail Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-avail")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "201", "triple", 150.0, 3)
	user1 := createTestUserWithName(t, db, "gh-u1", "u1@test.com", "User One", "public")
	user2 := createTestUserWithName(t, db, "gh-u2", "u2@test.com", "User Two", "public")

	createTestBooking(t, db, room.ID, user1.ID, conf.ID, models.BookingStatusConfirmed, "public")
	createTestBooking(t, db, room.ID, user2.ID, conf.ID, models.BookingStatusConfirmed, "public")

	resp, err := http.Get(srv.URL + "/conferences/conf-avail/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 1)

	assert.Equal(t, float64(2), body[0]["spots_taken"])
	assert.Equal(t, float64(1), body[0]["spots_available"])
}

func TestGetRooms_PublicOccupant_IncludesUserInfo(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-pub", "Public Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-pub")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "301", "double", 120.0, 2)
	user := createTestUserWithName(t, db, "gh-pub", "pub@test.com", "Public Alice", "public")
	createTestBooking(t, db, room.ID, user.ID, conf.ID, models.BookingStatusConfirmed, "public")

	resp, err := http.Get(srv.URL + "/conferences/conf-pub/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 1)

	occupants := body[0]["occupants"].([]interface{})
	require.Len(t, occupants, 1)

	occ := occupants[0].(map[string]interface{})
	assert.Equal(t, "Public Alice", occ["display_name"])
	assert.NotNil(t, occ["user_id"])
}

func TestGetRooms_PrivateOccupant_HidesUserInfo(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-priv", "Private Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-priv")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "401", "double", 120.0, 2)
	user := createTestUserWithName(t, db, "gh-priv", "priv@test.com", "Secret Bob", "public")
	createTestBooking(t, db, room.ID, user.ID, conf.ID, models.BookingStatusConfirmed, "private")

	resp, err := http.Get(srv.URL + "/conferences/conf-priv/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 1)

	occupants := body[0]["occupants"].([]interface{})
	require.Len(t, occupants, 1)

	occ := occupants[0].(map[string]interface{})
	assert.Equal(t, "Private attendee", occ["display_name"])
	assert.Nil(t, occ["user_id"])
}

func TestGetRooms_NonExistentSlug_Returns404(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/conferences/no-such-conf/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	errObj := body["error"].(map[string]interface{})
	assert.Equal(t, "not_found", errObj["code"])
}

func TestGetRooms_NoAuthRequired(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-noauth", "No Auth Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	// No Authorization header — should still return 200
	resp, err := http.Get(srv.URL + "/conferences/conf-noauth/rooms")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPostRooms_OrganizerOnly(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-post-room", "Post Room",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-post-room")
	require.NoError(t, err)

	organizer := createTestUserWithName(t, db, "gh-org-room", "org-room@test.com", "Organizer", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences/conf-post-room/rooms", strings.NewReader(`{"room_number":"901","room_type":"double","price_per_night":120,"capacity":2}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPostRooms_NonOrganizer_Returns403(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-post-room-forbidden", "Post Room Forbidden",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	nonOrganizer := createTestUserWithName(t, db, "gh-non-org-room", "non-org-room@test.com", "Non Organizer", "public")
	accessToken, _ := createTestSession(t, db, nonOrganizer.ID, tokenService)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences/conf-post-room-forbidden/rooms", strings.NewReader(`{"room_number":"901","room_type":"double","price_per_night":120,"capacity":2}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteRooms_FailsWhenBookingsExist(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-delete-room", "Delete Room",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-delete-room")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "902", "double", 120, 2)

	organizer := createTestUserWithName(t, db, "gh-org-delete-room", "org-delete-room@test.com", "Organizer", "public")
	attendee := createTestUserWithName(t, db, "gh-att-delete-room", "att-delete-room@test.com", "Attendee", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	createTestBooking(t, db, room.ID, attendee.ID, conf.ID, models.BookingStatusConfirmed, "public")
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/conferences/conf-delete-room/rooms/902", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetConferences_ReturnsRealAttendeeCount(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-attend", "Attendee Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-attend")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "501", "double", 120.0, 2)
	user1 := createTestUserWithName(t, db, "gh-att1", "att1@test.com", "Attendee 1", "public")
	user2 := createTestUserWithName(t, db, "gh-att2", "att2@test.com", "Attendee 2", "public")

	createTestBooking(t, db, room.ID, user1.ID, conf.ID, models.BookingStatusConfirmed, "public")
	createTestBooking(t, db, room.ID, user2.ID, conf.ID, models.BookingStatusConfirmed, "public")

	resp, err := http.Get(srv.URL + "/conferences")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	// Find our conference
	var found bool
	for _, c := range body {
		if c["slug"] == "conf-attend" {
			assert.Equal(t, float64(2), c["attendee_count"])
			found = true
			break
		}
	}
	assert.True(t, found, "conference conf-attend not found in response")
}

func TestGetAttendees_RequiresAuth(t *testing.T) {
	srv, _, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-attendees-noauth", "Attendees No Auth",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))

	resp, err := http.Get(srv.URL + "/conferences/conf-attendees-noauth/attendees")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetAttendees_NonOrganizerReturnsPublicRowsAndPrivateCount(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-attendees", "Attendees Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-attendees")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "601", "double", 120.0, 2)
	publicUser := createTestUserWithName(t, db, "gh-public", "public@test.com", "Public Alice", "public")
	privateUser := createTestUserWithName(t, db, "gh-private", "private@test.com", "Private Bob", "private")
	authUser := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, authUser.ID, tokenService)

	createTestBooking(t, db, room.ID, publicUser.ID, conf.ID, models.BookingStatusConfirmed, "public")
	createTestBooking(t, db, room.ID, privateUser.ID, conf.ID, models.BookingStatusConfirmed, "private")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-attendees/attendees", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, float64(1), body["private_attendees_count"])

	attendees := body["attendees"].([]interface{})
	require.Len(t, attendees, 1)
	row := attendees[0].(map[string]interface{})
	assert.Equal(t, "Public Alice", row["display_name"])
	roomInfo := row["room"].(map[string]interface{})
	assert.Equal(t, "601", roomInfo["room_number"])
}

func TestGetAttendees_OrganizerSeesPrivateRows(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-attendees-organizer", "Attendees Organizer",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-attendees-organizer")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "602", "double", 120.0, 2)
	privateUser := createTestUserWithName(t, db, "gh-private-visible", "private-visible@test.com", "Private Visible", "private")
	authUser := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, authUser.ID, tokenService)
	createTestOrganizer(t, db, conf.ID, authUser.ID, models.ConferenceOrganizerRoleAdmin)

	createTestBooking(t, db, room.ID, privateUser.ID, conf.ID, models.BookingStatusConfirmed, "private")

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-attendees-organizer/attendees", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, float64(0), body["private_attendees_count"])
	attendees := body["attendees"].([]interface{})
	require.Len(t, attendees, 1)
	row := attendees[0].(map[string]interface{})
	assert.Equal(t, "Private Visible", row["display_name"])
}

func TestGetAttendees_ConferenceNotFound(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	authUser := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, authUser.ID, tokenService)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/missing/attendees", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDashboard_OrganizerGetsMetricsAndPrivateData(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-dashboard", "Dashboard Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-dashboard")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "701", "double", 120.0, 2)
	privateUser := createTestUserWithName(t, db, "gh-private-dashboard", "private-dashboard@test.com", "Private Person", "private")
	organizer := createTestUserWithName(t, db, "gh-org-dashboard", "org-dashboard@test.com", "Org User", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	bookingRepo := sqlite.NewBookingRepository(db)
	_, err = bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         room.ID,
		UserID:         privateUser.ID,
		ConferenceID:   conf.ID,
		Status:         models.BookingStatusConfirmed,
		PrivacySetting: "private",
		Notes:          "Wheelchair access",
	})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-dashboard/dashboard?has_special_requests=true&q=private", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "conf-dashboard", body["conference_slug"])
	assert.Equal(t, float64(1), body["total_registrations"])
	attendees := body["attendees"].([]interface{})
	require.Len(t, attendees, 1)
	row := attendees[0].(map[string]interface{})
	assert.Equal(t, "Private Person", row["name"])
	assert.Equal(t, "private-dashboard@test.com", row["email"])
	assert.Equal(t, "private", row["privacy_setting"])
	assert.Equal(t, "Wheelchair access", row["dietary_accessibility_note"])
}

func TestGetDashboard_NonOrganizerReturns403(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-dashboard-403", "Dashboard 403",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-dashboard-403/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetExport_OrganizerGetsCSV_DefaultExcludesCancelled(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-export", "Export Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-export")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "801", "double", 120.0, 2)
	organizer := createTestUserWithName(t, db, "gh-org-export", "org-export@test.com", "Org Export", "public")
	activeUser := createTestUserWithName(t, db, "gh-export-active", "active@test.com", "Active User", "public")
	cancelledUser := createTestUserWithName(t, db, "gh-export-cancelled", "cancelled@test.com", "Cancelled User", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	bookingRepo := sqlite.NewBookingRepository(db)
	_, err = bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         room.ID,
		UserID:         activeUser.ID,
		ConferenceID:   conf.ID,
		Status:         models.BookingStatusConfirmed,
		PrivacySetting: "public",
		Notes:          "Vegan",
	})
	require.NoError(t, err)
	_, err = bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         room.ID,
		UserID:         cancelledUser.ID,
		ConferenceID:   conf.ID,
		Status:         models.BookingStatusCancelled,
		PrivacySetting: "public",
		Notes:          "No peanuts",
	})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-export/export", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/csv")
	body, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	output := string(body)
	assert.Contains(t, output, "name,email,room,dates,dietary/special notes")
	assert.Contains(t, output, "Active User,active@test.com,801")
	assert.NotContains(t, output, "Cancelled User")
}

func TestGetExport_IncludeCancelledIncludesCancelledRows(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-export-all", "Export All Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-export-all")
	require.NoError(t, err)

	room := createTestRoom(t, db, conf.ID, "802", "double", 120.0, 2)
	organizer := createTestUserWithName(t, db, "gh-org-export-all", "org-export-all@test.com", "Org Export All", "public")
	cancelledUser := createTestUserWithName(t, db, "gh-export-all-cancelled", "all-cancelled@test.com", "Cancelled User", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	bookingRepo := sqlite.NewBookingRepository(db)
	_, err = bookingRepo.Create(context.Background(), &models.Booking{
		RoomID:         room.ID,
		UserID:         cancelledUser.ID,
		ConferenceID:   conf.ID,
		Status:         models.BookingStatusCancelled,
		PrivacySetting: "public",
		Notes:          "Late cancel",
	})
	require.NoError(t, err)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-export-all/export?include_cancelled=true", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, readErr := io.ReadAll(resp.Body)
	require.NoError(t, readErr)
	assert.Contains(t, string(body), "Cancelled User,all-cancelled@test.com,802")
}

func TestGetExport_NonOrganizerReturns403(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-export-403", "Export 403",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/conferences/conf-export-403/export", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPostConferences_CreatesConferenceAndOwner(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	user := createTestUser(t, db)
	createTestConference(t, db, "conf-existing", "Existing",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	existingConf, err := confRepo.GetBySlug(context.Background(), "conf-existing")
	require.NoError(t, err)
	createTestOrganizer(t, db, existingConf.ID, user.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"slug":"conf-create-owner","name":"Create Owner","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	createdConf, err := confRepo.GetBySlug(context.Background(), "conf-create-owner")
	require.NoError(t, err)

	organizerRepo := sqlite.NewOrganizerRepository(db)
	membership, err := organizerRepo.GetByConferenceAndUser(context.Background(), createdConf.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, models.ConferenceOrganizerRoleOwner, membership.Role)
}

func TestPostConferences_BootstrapAllowedWithoutOrganizerMembership(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"slug":"conf-no-org","name":"No Organizer","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPostConferences_NonOrganizerWithExistingConference_Returns403(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-existing-no-org", "Existing",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"slug":"conf-no-org-2","name":"No Organizer","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestPostConferences_InvalidSlug_Returns400(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"slug":"Bad.Slug","name":"Invalid Slug","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPostConferences_SlugIsNormalizedServerSide(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	user := createTestUser(t, db)
	accessToken, _ := createTestSession(t, db, user.ID, tokenService)

	reqBody := `{"slug":"  My_Conf 2026  ","name":"Normalized","description":"desc","location":"Berlin","start_date":"2026-10-07","end_date":"2026-10-10","capacity":120}`
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	confRepo := sqlite.NewConferenceRepository(db)
	created, err := confRepo.GetBySlug(context.Background(), "my-conf-2026")
	require.NoError(t, err)
	assert.Equal(t, "my-conf-2026", created.Slug)
}

func TestPutConferences_OrganizerCanUpdate(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)

	createTestConference(t, db, "conf-edit", "Before Edit",
		time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-edit")
	require.NoError(t, err)

	organizer := createTestUserWithName(t, db, "gh-edit", "edit@test.com", "Edit User", "public")
	createTestOrganizer(t, db, conf.ID, organizer.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, organizer.ID, tokenService)

	reqBody := `{"name":"After Edit","description":"updated","location":"Munich","start_date":"2026-10-08","end_date":"2026-10-11","capacity":180}`
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/conferences/conf-edit", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updated, err := confRepo.GetBySlug(context.Background(), "conf-edit")
	require.NoError(t, err)
	assert.Equal(t, "After Edit", updated.Name)
	assert.Equal(t, "Munich", updated.Location)
	assert.Equal(t, 180, updated.Capacity)
}

func TestPostConferenceOrganizers_OwnerCanAdd(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-org-add", "Org Add",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-org-add")
	require.NoError(t, err)

	owner := createTestUserWithName(t, db, "gh-owner-add", "owner-add@test.com", "Owner Add", "public")
	target := createTestUserWithName(t, db, "gh-target-add", "target-add@test.com", "Target Add", "public")
	createTestOrganizer(t, db, conf.ID, owner.ID, models.ConferenceOrganizerRoleOwner)
	accessToken, _ := createTestSession(t, db, owner.ID, tokenService)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences/conf-org-add/organizers", strings.NewReader(`{"user_id":`+jsonNumber(target.ID)+`,"role":"admin"}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestDeleteConferenceOrganizers_OwnerCanRemove(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-org-remove", "Org Remove",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-org-remove")
	require.NoError(t, err)

	owner := createTestUserWithName(t, db, "gh-owner-remove", "owner-remove@test.com", "Owner Remove", "public")
	target := createTestUserWithName(t, db, "gh-target-remove", "target-remove@test.com", "Target Remove", "public")
	createTestOrganizer(t, db, conf.ID, owner.ID, models.ConferenceOrganizerRoleOwner)
	createTestOrganizer(t, db, conf.ID, target.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, owner.ID, tokenService)

	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/conferences/conf-org-remove/organizers/"+jsonNumber(target.ID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestPostConferenceOrganizers_AdminForbidden(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	createTestConference(t, db, "conf-org-admin", "Org Admin",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-org-admin")
	require.NoError(t, err)

	admin := createTestUserWithName(t, db, "gh-admin-forbid", "admin-forbid@test.com", "Admin Forbid", "public")
	target := createTestUserWithName(t, db, "gh-target-forbid", "target-forbid@test.com", "Target Forbid", "public")
	createTestOrganizer(t, db, conf.ID, admin.ID, models.ConferenceOrganizerRoleAdmin)
	accessToken, _ := createTestSession(t, db, admin.ID, tokenService)

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/conferences/conf-org-admin/organizers", strings.NewReader(`{"user_id":`+jsonNumber(target.ID)+`,"role":"admin"}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func createTestRoommateRequest(t *testing.T, db *sql.DB, requesterID, targetID, roomID int64) *models.RoommateRequest {
	t.Helper()
	repo := sqlite.NewRoommateRequestRepository(db)
	req, err := repo.Create(context.Background(), &models.RoommateRequest{
		RequesterID: requesterID,
		TargetID:    targetID,
		RoomID:      roomID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)
	return req
}

func TestRequests_RequireAuth(t *testing.T) {
	srv, _, _ := setupIntegrationRouter(t)

	resp, err := http.Get(srv.URL + "/requests")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPostRequests_CreatesRoommateRequest(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	requester := createTestUserWithName(t, db, "gh-req-requester", "req-requester@test.com", "Requester", "public")
	target := createTestUserWithName(t, db, "gh-req-target", "req-target@test.com", "Target", "public")
	accessToken, _ := createTestSession(t, db, requester.ID, tokenService)

	createTestConference(t, db, "conf-req-create", "Requests Create",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-req-create")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "701", "double", 120, 2)
	createTestBooking(t, db, room.ID, requester.ID, conf.ID, models.BookingStatusConfirmed, "public")

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/requests", strings.NewReader(`{"target_id":`+jsonNumber(target.ID)+`,"room_id":`+jsonNumber(room.ID)+`}`))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestGetRequests_ReturnsIncomingAndOutgoing(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	requester := createTestUserWithName(t, db, "gh-req-list-requester", "req-list-requester@test.com", "Requester", "public")
	target := createTestUserWithName(t, db, "gh-req-list-target", "req-list-target@test.com", "Target", "public")
	third := createTestUserWithName(t, db, "gh-req-list-third", "req-list-third@test.com", "Third", "public")
	accessToken, _ := createTestSession(t, db, requester.ID, tokenService)

	createTestConference(t, db, "conf-req-list", "Requests List",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-req-list")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "702", "double", 120, 3)
	createTestBooking(t, db, room.ID, requester.ID, conf.ID, models.BookingStatusConfirmed, "public")
	createTestBooking(t, db, room.ID, third.ID, conf.ID, models.BookingStatusConfirmed, "public")

	createTestRoommateRequest(t, db, requester.ID, target.ID, room.ID)
	createTestRoommateRequest(t, db, third.ID, requester.ID, room.ID)

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/requests", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.Len(t, body, 2)

	directions := map[string]bool{}
	for _, row := range body {
		directions[row["direction"].(string)] = true
		assert.NotEmpty(t, row["requester_name"])
		assert.NotEmpty(t, row["target_name"])
		assert.NotEmpty(t, row["room_number"])
		assert.NotEmpty(t, row["room_type"])
		assert.NotZero(t, row["conference_id"])
	}
	assert.True(t, directions["incoming"])
	assert.True(t, directions["outgoing"])
}

func TestPutRequestsAccept_AddsTargetToRoom(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	requester := createTestUserWithName(t, db, "gh-req-acc-requester", "req-acc-requester@test.com", "Requester", "public")
	target := createTestUserWithName(t, db, "gh-req-acc-target", "req-acc-target@test.com", "Target", "public")

	createTestConference(t, db, "conf-req-accept", "Requests Accept",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-req-accept")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "703", "double", 120, 2)
	createTestBooking(t, db, room.ID, requester.ID, conf.ID, models.BookingStatusConfirmed, "public")
	request := createTestRoommateRequest(t, db, requester.ID, target.ID, room.ID)

	targetToken, _ := createTestSession(t, db, target.ID, tokenService)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/requests/"+jsonNumber(request.ID)+"/accept", nil)
	req.Header.Set("Authorization", "Bearer "+targetToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	reqRepo := sqlite.NewRoommateRequestRepository(db)
	updated, err := reqRepo.GetByID(context.Background(), request.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusAccepted, updated.Status)

	bookingRepo := sqlite.NewBookingRepository(db)
	targetBooking, err := bookingRepo.GetActiveByUserAndConference(context.Background(), target.ID, conf.ID)
	require.NoError(t, err)
	assert.Equal(t, room.ID, targetBooking.RoomID)
}

func TestPutRequestsDecline_UpdatesStatusOnly(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	requester := createTestUserWithName(t, db, "gh-req-dec-requester", "req-dec-requester@test.com", "Requester", "public")
	target := createTestUserWithName(t, db, "gh-req-dec-target", "req-dec-target@test.com", "Target", "public")

	createTestConference(t, db, "conf-req-decline", "Requests Decline",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-req-decline")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "704", "double", 120, 2)
	createTestBooking(t, db, room.ID, requester.ID, conf.ID, models.BookingStatusConfirmed, "public")
	request := createTestRoommateRequest(t, db, requester.ID, target.ID, room.ID)

	targetToken, _ := createTestSession(t, db, target.ID, tokenService)
	req, _ := http.NewRequest(http.MethodPut, srv.URL+"/requests/"+jsonNumber(request.ID)+"/decline", nil)
	req.Header.Set("Authorization", "Bearer "+targetToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	reqRepo := sqlite.NewRoommateRequestRepository(db)
	updated, err := reqRepo.GetByID(context.Background(), request.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusDeclined, updated.Status)

	bookingRepo := sqlite.NewBookingRepository(db)
	_, err = bookingRepo.GetActiveByUserAndConference(context.Background(), target.ID, conf.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)
}

func TestDeleteBookings_CancelsBookingAndPreservesRoommate(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	canceller := createTestUserWithName(t, db, "gh-cancel-owner", "cancel-owner@test.com", "Canceller", "public")
	roommate := createTestUserWithName(t, db, "gh-cancel-roommate", "cancel-roommate@test.com", "Roommate", "public")
	target := createTestUserWithName(t, db, "gh-cancel-target", "cancel-target@test.com", "Target", "public")

	createTestConference(t, db, "conf-cancel", "Cancel Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-cancel")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "204", "double", 120, 2)
	bookingToCancel := createTestBooking(t, db, room.ID, canceller.ID, conf.ID, models.BookingStatusConfirmed, "public")
	createTestBooking(t, db, room.ID, roommate.ID, conf.ID, models.BookingStatusConfirmed, "public")
	outgoing := createTestRoommateRequest(t, db, canceller.ID, target.ID, room.ID)

	token, _ := createTestSession(t, db, canceller.ID, tokenService)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/bookings/"+jsonNumber(bookingToCancel.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bookingRepo := sqlite.NewBookingRepository(db)
	cancelledBooking, err := bookingRepo.GetByID(context.Background(), bookingToCancel.ID)
	require.NoError(t, err)
	assert.Equal(t, models.BookingStatusCancelled, cancelledBooking.Status)
	require.NotNil(t, cancelledBooking.CancelledAt)

	roomBookings, err := bookingRepo.ListByRoom(context.Background(), room.ID)
	require.NoError(t, err)
	require.Len(t, roomBookings, 1)
	assert.Equal(t, roommate.ID, roomBookings[0].UserID)
	_, err = bookingRepo.GetActiveByUserAndConference(context.Background(), target.ID, conf.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrBookingNotFound)

	requestRepo := sqlite.NewRoommateRequestRepository(db)
	updatedOutgoing, err := requestRepo.GetByID(context.Background(), outgoing.ID)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusCancelled, updatedOutgoing.Status)

	bookingsResp, err := http.NewRequest(http.MethodGet, srv.URL+"/bookings", nil)
	require.NoError(t, err)
	bookingsResp.Header.Set("Authorization", "Bearer "+token)
	bookingsHTTPResp, err := http.DefaultClient.Do(bookingsResp)
	require.NoError(t, err)
	defer func() { _ = bookingsHTTPResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, bookingsHTTPResp.StatusCode)

	var activeBookings []map[string]interface{}
	require.NoError(t, json.NewDecoder(bookingsHTTPResp.Body).Decode(&activeBookings))
	assert.Empty(t, activeBookings)

	roomResp, err := http.Get(srv.URL + "/conferences/conf-cancel/rooms")
	require.NoError(t, err)
	defer func() { _ = roomResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, roomResp.StatusCode)

	var roomBody []map[string]interface{}
	require.NoError(t, json.NewDecoder(roomResp.Body).Decode(&roomBody))
	require.Len(t, roomBody, 1)
	assert.Equal(t, "204", roomBody[0]["room_number"])
	assert.Equal(t, float64(1), roomBody[0]["spots_taken"])
	assert.Equal(t, float64(1), roomBody[0]["spots_available"])
}

func TestDeleteBookings_ForbidCancellingOthersBooking(t *testing.T) {
	srv, tokenService, db := setupIntegrationRouter(t)
	owner := createTestUserWithName(t, db, "gh-cancel-owner2", "cancel-owner2@test.com", "Owner", "public")
	otherUser := createTestUserWithName(t, db, "gh-cancel-other2", "cancel-other2@test.com", "Other", "public")

	createTestConference(t, db, "conf-cancel-forbidden", "Cancel Forbidden Conf",
		time.Now().Add(30*24*time.Hour), time.Now().Add(33*24*time.Hour))
	confRepo := sqlite.NewConferenceRepository(db)
	conf, err := confRepo.GetBySlug(context.Background(), "conf-cancel-forbidden")
	require.NoError(t, err)
	room := createTestRoom(t, db, conf.ID, "303", "double", 120, 2)
	ownersBooking := createTestBooking(t, db, room.ID, owner.ID, conf.ID, models.BookingStatusConfirmed, "public")

	otherToken, _ := createTestSession(t, db, otherUser.ID, tokenService)
	req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/bookings/"+jsonNumber(ownersBooking.ID), nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func jsonNumber(v int64) string {
	return strconv.FormatInt(v, 10)
}
