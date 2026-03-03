package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := NewConnectionManager(context.Background(), ":memory:")
	require.NoError(t, err)

	err = RunMigrations(db)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, db.Close())
	})

	return db
}

func TestUserRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "gh_1",
		Email:       "user1@example.com",
		DisplayName: "User One",
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "gh_1", created.GitHubID)
	assert.Equal(t, "public", created.PrivacySetting)
	assert.False(t, created.CreatedAt.IsZero())
	assert.False(t, created.UpdatedAt.IsZero())
}

func TestUserRepository_Create_DuplicateGitHubID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "duplicate_gh",
		Email:       "first@example.com",
		DisplayName: "First",
	})
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), &models.User{
		GitHubID:    "duplicate_gh",
		Email:       "second@example.com",
		DisplayName: "Second",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserExists))
}

func TestUserRepository_GetByID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "by_id",
		Email:       "byid@example.com",
		DisplayName: "By ID",
	})
	require.NoError(t, err)

	fetched, err := repo.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "by_id", fetched.GitHubID)
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByID(context.Background(), 99999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))
}

func TestUserRepository_GetByGitHubID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "gh_lookup",
		Email:       "lookup@example.com",
		DisplayName: "Lookup",
	})
	require.NoError(t, err)

	fetched, err := repo.GetByGitHubID(context.Background(), created.GitHubID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Email, fetched.Email)
}

func TestUserRepository_GetByGitHubID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByGitHubID(context.Background(), "missing_gh")
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))
}

func TestUserRepository_Update_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "update_target",
		Email:       "before@example.com",
		DisplayName: "Before",
	})
	require.NoError(t, err)

	created.Email = "after@example.com"
	created.DisplayName = "After"
	created.PrivacySetting = "private"

	updated, err := repo.Update(context.Background(), created)
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "update_target", updated.GitHubID)
	assert.Equal(t, "after@example.com", updated.Email)
	assert.Equal(t, "After", updated.DisplayName)
	assert.Equal(t, "private", updated.PrivacySetting)
}

func TestUserRepository_Update_CannotChangeID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	created, err := repo.Create(context.Background(), &models.User{
		GitHubID:    "immutable",
		Email:       "immutable@example.com",
		DisplayName: "Immutable",
	})
	require.NoError(t, err)

	immutableID := created.ID
	created.ID = immutableID + 10
	created.Email = "attempt@example.com"

	_, err = repo.Update(context.Background(), created)
	require.Error(t, err)
	assert.True(t, errors.Is(err, repository.ErrUserNotFound))

	fetched, err := repo.GetByID(context.Background(), immutableID)
	require.NoError(t, err)
	assert.Equal(t, "immutable@example.com", fetched.Email)
}
