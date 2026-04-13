package sqlite

import (
	"context"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type requestFixture struct {
	repo     *RoommateRequestRepository
	confRepo *ConferenceRepository
	roomRepo *RoomRepository
	userRepo *UserRepository
	conf     *models.Conference
	room     *models.Room
	alice    *models.User
	bob      *models.User
}

func setupRequestFixture(t *testing.T) *requestFixture {
	t.Helper()
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	roomRepo := NewRoomRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewRoommateRequestRepository(db)

	conf, err := confRepo.Create(context.Background(), newTestConference("request-conf"))
	require.NoError(t, err)
	room, err := roomRepo.Create(context.Background(), newTestRoom(conf.ID, "201"))
	require.NoError(t, err)

	alice, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "gh_alice_req",
		Email:       "alice@req.test",
		DisplayName: "Alice",
	})
	require.NoError(t, err)
	bob, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "gh_bob_req",
		Email:       "bob@req.test",
		DisplayName: "Bob",
	})
	require.NoError(t, err)

	return &requestFixture{
		repo:     repo,
		confRepo: confRepo,
		roomRepo: roomRepo,
		userRepo: userRepo,
		conf:     conf,
		room:     room,
		alice:    alice,
		bob:      bob,
	}
}

func TestRoommateRequestRepository_CreateAndGetByID(t *testing.T) {
	f := setupRequestFixture(t)

	created, err := f.repo.Create(context.Background(), &models.RoommateRequest{
		RequesterID: f.alice.ID,
		TargetID:    f.bob.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, models.RoommateRequestStatusPending, created.Status)

	found, err := f.repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, f.alice.ID, found.RequesterID)
	assert.Equal(t, f.bob.ID, found.TargetID)
}

func TestRoommateRequestRepository_CreateDuplicatePendingFails(t *testing.T) {
	f := setupRequestFixture(t)
	ctx := context.Background()

	_, err := f.repo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.alice.ID,
		TargetID:    f.bob.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	_, err = f.repo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.alice.ID,
		TargetID:    f.bob.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrRoommateRequestExists)
}

func TestRoommateRequestRepository_ListByUser(t *testing.T) {
	f := setupRequestFixture(t)
	ctx := context.Background()

	_, err := f.repo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.alice.ID,
		TargetID:    f.bob.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	thirdUser, err := f.userRepo.Create(ctx, &models.User{
		GitHubID:    "gh_charlie_req",
		Email:       "charlie@req.test",
		DisplayName: "Charlie",
	})
	require.NoError(t, err)

	_, err = f.repo.Create(ctx, &models.RoommateRequest{
		RequesterID: thirdUser.ID,
		TargetID:    f.alice.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	rows, err := f.repo.ListByUser(ctx, f.alice.ID)
	require.NoError(t, err)
	assert.Len(t, rows, 2)
}

func TestRoommateRequestRepository_UpdateStatus(t *testing.T) {
	f := setupRequestFixture(t)
	ctx := context.Background()

	created, err := f.repo.Create(ctx, &models.RoommateRequest{
		RequesterID: f.alice.ID,
		TargetID:    f.bob.ID,
		RoomID:      f.room.ID,
		Status:      models.RoommateRequestStatusPending,
	})
	require.NoError(t, err)

	updated, err := f.repo.UpdateStatus(ctx, created.ID, models.RoommateRequestStatusAccepted)
	require.NoError(t, err)
	assert.Equal(t, models.RoommateRequestStatusAccepted, updated.Status)
}
