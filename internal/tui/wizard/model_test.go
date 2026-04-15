package wizard

import (
	"context"
	"database/sql"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/katurdays/unconf/internal/api"
	"github.com/katurdays/unconf/internal/api/handlers"
	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/client"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository/sqlite"
	"github.com/katurdays/unconf/internal/service"
	"github.com/katurdays/unconf/internal/tui/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const wizardTestSymmetricKey = "0123456789abcdef0123456789abcdef"

func testRoomSelection() RoomSelection {
	return RoomSelection{
		ConferenceSlug: "socrates-26",
		Room: client.RoomResponse{
			ID:             12,
			ConferenceID:   4,
			RoomNumber:     "205",
			RoomType:       "double",
			PricePerNight:  149,
			SpotsAvailable: 1,
			Capacity:       2,
			Occupants: []client.RoomOccupantResponse{
				{DisplayName: "Alex"},
			},
		},
	}
}

type wizardIntegrationHarness struct {
	server       *httptest.Server
	db           *sql.DB
	tokenService *auth.TokenService
}

func newWizardIntegrationHarness(t *testing.T) *wizardIntegrationHarness {
	t.Helper()

	db, err := sqlite.NewConnectionManager(context.Background(), ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, sqlite.RunMigrations(db))

	tokenService, err := auth.NewTokenService(wizardTestSymmetricKey)
	require.NoError(t, err)

	conferenceRepo := sqlite.NewConferenceRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	bookingRepo := sqlite.NewBookingRepository(db)
	userRepo := sqlite.NewUserRepository(db)

	bookingHandler := handlers.NewBookingHandler(service.NewBookingService(bookingRepo, roomRepo, conferenceRepo, userRepo))
	router := api.NewRouter(nil, tokenService, nil, nil, nil, nil, nil, nil, bookingHandler, nil)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	return &wizardIntegrationHarness{server: server, db: db, tokenService: tokenService}
}

func createWizardIntegrationUser(t *testing.T, db *sql.DB, githubID, email, displayName, privacy string) *models.User {
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

func createWizardIntegrationConference(t *testing.T, db *sql.DB, slug string) *models.Conference {
	t.Helper()

	conferenceRepo := sqlite.NewConferenceRepository(db)
	conference, err := conferenceRepo.Create(context.Background(), &models.Conference{
		Slug:        slug,
		Name:        "SoCraTes 2026",
		Description: "Wizard integration conference",
		Location:    "Bergamo",
		StartDate:   time.Now().Add(30 * 24 * time.Hour).UTC(),
		EndDate:     time.Now().Add(33 * 24 * time.Hour).UTC(),
		Capacity:    120,
	})
	require.NoError(t, err)

	return conference
}

func createWizardIntegrationRoom(t *testing.T, db *sql.DB, conferenceID int64, roomNumber string) *models.Room {
	t.Helper()

	roomRepo := sqlite.NewRoomRepository(db)
	room, err := roomRepo.Create(context.Background(), &models.Room{
		ConferenceID:  conferenceID,
		RoomNumber:    roomNumber,
		RoomType:      "double",
		PricePerNight: 149,
		Capacity:      2,
	})
	require.NoError(t, err)

	return room
}

func createWizardIntegrationSession(t *testing.T, db *sql.DB, userID int64, tokenService *auth.TokenService) string {
	t.Helper()

	refreshRepo := sqlite.NewRefreshSessionRepository(db)
	now := time.Now().UTC()

	session, err := refreshRepo.Create(context.Background(), &models.RefreshSession{
		UserID:        userID,
		TokenHash:     "wizard-test-hash",
		ExpiresAt:     now.Add(7 * 24 * time.Hour),
		IssuedAt:      now,
		LastAccessJTI: "wizard-test-jti",
	})
	require.NoError(t, err)

	accessToken, err := tokenService.IssueAccessToken(context.Background(), auth.AccessTokenInput{
		UserID:     userID,
		SessionID:  session.ID,
		Issuer:     "unconf-api",
		Audience:   "unconf-cli",
		NotBefore:  now,
		IssuedAt:   now,
		JTI:        "wizard-test-access-jti",
		ExpiryTime: now.Add(1 * time.Hour),
	})
	require.NoError(t, err)

	return accessToken
}

func TestModel_StepProgressionAndBackNavigation(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepPrivacy, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, StepReview, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, StepNotes, next.(*Model).step)
}

func TestModel_PrivacyAndNotesStatePersistedIntoPayload(t *testing.T) {
	var captured client.CreateBookingRequest
	model := NewModel(context.Background(), testRoomSelection(), func(_ context.Context, input client.CreateBookingRequest) (*client.BookingResponse, error) {
		captured = input
		return &client.BookingResponse{ID: 99, RoomID: input.RoomID, ConferenceID: input.ConferenceID, Status: "requested", PrivacySetting: input.PrivacySetting}, nil
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, privacyPrivate, next.(*Model).privacySetting)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dietary")})
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Nil(t, cmd)

	review := next.(*Model)
	require.Equal(t, StepReview, review.step)

	next, cmd = review.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	next, _ = next.(*Model).Update(msg)

	final := next.(*Model)
	assert.Equal(t, StepSuccess, final.step)
	assert.Equal(t, int64(12), captured.RoomID)
	assert.Equal(t, int64(4), captured.ConferenceID)
	assert.Equal(t, privacyPrivate, captured.PrivacySetting)
	assert.Equal(t, "dietary", captured.Notes)
}

func TestModel_SubmitsBookingAgainstRunningServer(t *testing.T) {
	harness := newWizardIntegrationHarness(t)
	conference := createWizardIntegrationConference(t, harness.db, "socrates-26")
	room := createWizardIntegrationRoom(t, harness.db, conference.ID, "205")
	user := createWizardIntegrationUser(t, harness.db, "gh-wizard-book", "wizard-book@test.com", "Wizard Booker", "public")
	accessToken := createWizardIntegrationSession(t, harness.db, user.ID, harness.tokenService)

	store := auth.NewMockTokenStore()
	store.SetTokens(accessToken, "valid-refresh-token")
	authClient := client.NewAuthenticatedClient(client.NewClient(harness.server.URL), store)
	roomSelection := RoomSelection{
		ConferenceSlug: conference.Slug,
		Room: client.RoomResponse{
			ID:             room.ID,
			ConferenceID:   conference.ID,
			RoomNumber:     room.RoomNumber,
			RoomType:       room.RoomType,
			PricePerNight:  room.PricePerNight,
			SpotsAvailable: 1,
			Capacity:       room.Capacity,
		},
	}
	model := NewModel(context.Background(), roomSelection, authClient.CreateBooking, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRight})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("dietary")})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	next, _ = next.(*Model).Update(cmd())

	final := next.(*Model)
	require.Equal(t, StepSuccess, final.step)
	booking, ok := final.BookingResult()
	require.True(t, ok)
	assert.Equal(t, int64(room.ID), booking.RoomID)
	assert.Equal(t, conference.ID, booking.ConferenceID)
	assert.Equal(t, "requested", booking.Status)
	assert.Equal(t, privacyPrivate, booking.PrivacySetting)
	assert.Equal(t, "dietary", booking.Notes)

	bookingRepo := sqlite.NewBookingRepository(harness.db)
	createdBooking, err := bookingRepo.GetActiveByUserAndConference(context.Background(), user.ID, conference.ID)
	require.NoError(t, err)
	assert.Equal(t, room.ID, createdBooking.RoomID)
	assert.Equal(t, models.BookingStatusRequested, createdBooking.Status)
	assert.Equal(t, privacyPrivate, createdBooking.PrivacySetting)
	assert.Equal(t, "dietary", createdBooking.Notes)
}

func TestModel_SubmitErrorAllowsRetryAndBack(t *testing.T) {
	attempts := 0
	model := NewModel(context.Background(), testRoomSelection(), func(_ context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("room_full")
		}
		return &client.BookingResponse{ID: 1, Status: "requested"}, nil
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	next, _ = next.(*Model).Update(cmd())

	failed := next.(*Model)
	require.Error(t, failed.err)
	assert.Equal(t, StepReview, failed.step)

	next, _ = failed.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	assert.Equal(t, StepNotes, next.(*Model).step)
}

func TestModel_NotesStepAllowsTypingNavigationLetters(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bnb")})
	notes := next.(*Model)
	assert.Equal(t, StepNotes, notes.step)
	assert.Equal(t, "bnb", notes.notes)
}

func TestModel_NotesStepSingleLetterInputDoesNotNavigate(t *testing.T) {
	t.Run("n stays in notes step", func(t *testing.T) {
		model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.Equal(t, StepNotes, next.(*Model).step)

		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
		notes := next.(*Model)

		assert.Equal(t, StepNotes, notes.step)
		assert.Equal(t, "n", notes.notes)
	})

	t.Run("b stays in notes step", func(t *testing.T) {
		model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

		next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
		require.Equal(t, StepNotes, next.(*Model).step)

		next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
		notes := next.(*Model)

		assert.Equal(t, StepNotes, notes.step)
		assert.Equal(t, "b", notes.notes)
	})
}

func TestModel_NotesBackspaceRemovesLastRune(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("naive😄")})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyBackspace})
	notes := next.(*Model)

	assert.Equal(t, "naive", notes.notes)
	assert.Equal(t, StepNotes, notes.step)
}

func TestModel_NotesRuneLimit_AllowsMultibyteContentUnderRuneLimit(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	multibyte := strings.Repeat("😄", 200)
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(multibyte)})
	notes := next.(*Model)

	assert.Equal(t, 200, utf8.RuneCountInString(notes.notes))
	assert.Equal(t, multibyte, notes.notes)
}

func TestModel_NotesRuneLimit_CapsInputAtMaxRunes(t *testing.T) {
	model := NewModel(context.Background(), testRoomSelection(), nil, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.Equal(t, StepNotes, next.(*Model).step)

	overLimit := strings.Repeat("😄", notesMaxRunes+20)
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(overLimit)})
	notes := next.(*Model)

	assert.Equal(t, notesMaxRunes, utf8.RuneCountInString(notes.notes))
	assert.Equal(t, strings.Repeat("😄", notesMaxRunes), notes.notes)
}

func TestModel_SubmitCanceledContextShowsErrorState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	model := NewModel(ctx, testRoomSelection(), func(ctx context.Context, _ client.CreateBookingRequest) (*client.BookingResponse, error) {
		return nil, ctx.Err()
	}, common.NewStyles())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	next, _ = next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})

	cancel()
	next, cmd := next.(*Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	next, _ = next.(*Model).Update(cmd())

	failed := next.(*Model)
	require.Error(t, failed.err)
	assert.ErrorIs(t, failed.err, context.Canceled)
	assert.Contains(t, failed.err.Error(), "submit booking")
	assert.Equal(t, StepReview, failed.step)
	assert.False(t, failed.submitting)
}
