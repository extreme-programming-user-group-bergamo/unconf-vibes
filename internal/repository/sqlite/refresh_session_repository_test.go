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
		LastAccessJTI: "access-jti-1",
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	fetched, err := repo.GetByTokenHash(context.Background(), "hash_1")
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.UserID, fetched.UserID)
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
		LastAccessJTI: "jti-current",
	})
	require.NoError(t, err)

	replacement, err := repo.Rotate(context.Background(), current.ID, &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     "replacement-hash",
		ExpiresAt:     time.Now().UTC().Add(2 * time.Hour),
		IssuedAt:      time.Now().UTC(),
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
