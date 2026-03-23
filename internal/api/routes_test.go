package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/models"
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
	bookingRepo := sqlite.NewBookingRepository(db)
	conferenceService := service.NewConferenceService(conferenceRepo, bookingRepo)
	conferenceHandler := handlers.NewConferenceHandler(conferenceService)

	router := NewRouter(authHandler, tokenService, userHandler, conferenceHandler)
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
