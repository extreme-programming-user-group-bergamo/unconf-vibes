package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshSessionRepository_CreateAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshSessionRepository(db)

	created, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        seedUser(t, db, "refresh_create_get").ID,
		TokenHash:     "hash_1",
		ExpiresAt:     time.Now().UTC().Add(1 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		ClientInfo:    "ua=test-client",
		LastAccessJTI: "access-jti-1",
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	fetched, err := repo.GetByTokenHash(context.Background(), "hash_1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.UserID, fetched.UserID)
	assert.Equal(t, "ua=test-client", fetched.ClientInfo)
}

func TestRefreshSessionRepository_GetByTokenHashNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRefreshSessionRepository(db)

	_, err := repo.GetByTokenHash(context.Background(), "missing")
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrRefreshSessionNotFound))
}

func TestRefreshSessionRepository_Rotate(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db, "refresh_rotate")
	repo := NewRefreshSessionRepository(db)

	current, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "current-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		ClientInfo:    "ua=current",
		LastAccessJTI: "jti-current",
	})
	require.NoError(t, err)

	replacement, err := repo.Rotate(context.Background(), current.ID, &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "replacement-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		ClientInfo:    "ua=replacement",
		LastAccessJTI: "jti-replacement",
	})
	require.NoError(t, err)
	require.NotNil(t, replacement)

	updatedCurrent, err := repo.GetByTokenHash(context.Background(), "current-hash")
	require.NoError(t, err)
	require.NotNil(t, updatedCurrent.RevokedAt)
	require.NotNil(t, updatedCurrent.ReplacedByID)
	assert.Equal(t, replacement.ID, *updatedCurrent.ReplacedByID)
}

func TestRefreshSessionRepository_ListActiveByUser(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db, "refresh_list_active")
	repo := NewRefreshSessionRepository(db)

	_, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "active-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		ClientInfo:    "ua=active",
		LastAccessJTI: "jti-active",
	})
	require.NoError(t, err)

	expired, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "expired-hash",
		ExpiresAt:     time.Now().UTC().Add(-2 * time.Hour),
		IssuedAt:      time.Now().UTC().Add(-3 * time.Hour),
		ClientInfo:    "ua=expired",
		LastAccessJTI: "jti-expired",
	})
	require.NoError(t, err)
	require.NoError(t, repo.RevokeByID(context.Background(), expired.ID))

	active, err := repo.ListActiveByUser(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, active, 1)
	assert.Equal(t, "active-hash", active[0].TokenHash)
	assert.Equal(t, "ua=active", active[0].ClientInfo)
}

func TestRefreshSessionRepository_RevokeByUserAndID(t *testing.T) {
	db := setupTestDB(t)
	owner := seedUser(t, db, "refresh_owner")
	otherUser := seedUser(t, db, "refresh_other")
	repo := NewRefreshSessionRepository(db)

	owned, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        owner.ID,
		TokenHash:     "owned-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		LastAccessJTI: "jti-owned",
	})
	require.NoError(t, err)

	other, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        otherUser.ID,
		TokenHash:     "other-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		LastAccessJTI: "jti-other",
	})
	require.NoError(t, err)

	require.NoError(t, repo.RevokeByUserAndID(context.Background(), owner.ID, owned.ID))
	revokedOwned, err := repo.GetByTokenHash(context.Background(), "owned-hash")
	require.NoError(t, err)
	require.NotNil(t, revokedOwned.RevokedAt)

	err = repo.RevokeByUserAndID(context.Background(), owner.ID, other.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrRefreshSessionNotFound)
}

func TestRefreshSessionRepository_RevokeAllByUserExceptSession(t *testing.T) {
	db := setupTestDB(t)
	user := seedUser(t, db, "refresh_revoke_others")
	repo := NewRefreshSessionRepository(db)

	current, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "current-session-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		LastAccessJTI: "jti-current",
	})
	require.NoError(t, err)

	other, err := repo.Create(context.Background(), &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "other-session-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
		LastAccessJTI: "jti-other",
	})
	require.NoError(t, err)

	affected, err := repo.RevokeAllByUserExceptSession(context.Background(), user.ID, current.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), affected)

	currentSession, err := repo.GetByTokenHash(context.Background(), "current-session-hash")
	require.NoError(t, err)
	assert.Nil(t, currentSession.RevokedAt)

	otherSession, err := repo.GetByTokenHash(context.Background(), "other-session-hash")
	require.NoError(t, err)
	assert.Equal(t, other.ID, otherSession.ID)
	require.NotNil(t, otherSession.RevokedAt)
}

func seedUser(t *testing.T, db *sql.DB, githubID string) *models.User {
	t.Helper()

	userRepo := NewUserRepository(db)
	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    githubID,
		Email:       githubID + "@example.com",
		DisplayName: githubID,
	})
	require.NoError(t, err)

	return user
}
