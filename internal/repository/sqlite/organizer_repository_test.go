package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizerRepository_AddAndGet_Success(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-add-get",
		Name:        "Org Add",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "org-add-get-gh",
		Email:       "org-add-get@test.com",
		DisplayName: "Org User",
	})
	require.NoError(t, err)

	created, err := repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)
	assert.Equal(t, models.ConferenceOrganizerRoleAdmin, created.Role)

	fetched, err := repo.GetByConferenceAndUser(context.Background(), conf.ID, user.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, models.ConferenceOrganizerRoleAdmin, fetched.Role)
}

func TestOrganizerRepository_Add_DuplicateMembership(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-dup",
		Name:        "Org Dup",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "org-dup-gh",
		Email:       "org-dup@test.com",
		DisplayName: "Org Dup User",
	})
	require.NoError(t, err)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleOwner,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrConferenceOrganizerExists)
}

func TestOrganizerRepository_IsOrganizerAndRemove(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-remove",
		Name:        "Org Remove",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "org-remove-gh",
		Email:       "org-remove@test.com",
		DisplayName: "Org Remove User",
	})
	require.NoError(t, err)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)

	isOrganizer, err := repo.IsOrganizer(context.Background(), conf.ID, user.ID)
	require.NoError(t, err)
	assert.True(t, isOrganizer)

	require.NoError(t, repo.RemoveByConferenceAndUser(context.Background(), conf.ID, user.ID))

	isOrganizer, err = repo.IsOrganizer(context.Background(), conf.ID, user.ID)
	require.NoError(t, err)
	assert.False(t, isOrganizer)
}

func TestOrganizerRepository_ConferenceDeleteCascadesMemberships(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-cascade",
		Name:        "Org Cascade",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "org-cascade-gh",
		Email:       "org-cascade@test.com",
		DisplayName: "Org Cascade User",
	})
	require.NoError(t, err)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), "DELETE FROM conferences WHERE id = ?", conf.ID)
	require.NoError(t, err)

	_, err = repo.GetByConferenceAndUser(context.Background(), conf.ID, user.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrConferenceOrganizerNotFound)

	var count int
	err = db.QueryRowContext(context.Background(), "SELECT COUNT(1) FROM conference_organizers WHERE conference_id = ?", conf.ID).Scan(&count)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestOrganizerRepository_IsOrganizerForAnyConference(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-any",
		Name:        "Org Any",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	user, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "org-any-gh",
		Email:       "org-any@test.com",
		DisplayName: "Org Any User",
	})
	require.NoError(t, err)

	ok, err := repo.IsOrganizerForAnyConference(context.Background(), user.ID)
	require.NoError(t, err)
	assert.False(t, ok)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       user.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)

	ok, err = repo.IsOrganizerForAnyConference(context.Background(), user.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestOrganizerRepository_ListEmailsByConference(t *testing.T) {
	db := setupTestDB(t)
	confRepo := NewConferenceRepository(db)
	userRepo := NewUserRepository(db)
	repo := NewOrganizerRepository(db)

	conf, err := confRepo.Create(context.Background(), &models.Conference{
		Slug:        "org-email-list",
		Name:        "Org Email List",
		Description: "desc",
		Location:    "City",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)

	owner, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "owner-gh",
		Email:       "owner@example.com",
		DisplayName: "Owner",
	})
	require.NoError(t, err)
	admin, err := userRepo.Create(context.Background(), &models.User{
		GitHubID:    "admin-gh",
		Email:       "admin@example.com",
		DisplayName: "Admin",
	})
	require.NoError(t, err)

	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       admin.ID,
		Role:         models.ConferenceOrganizerRoleAdmin,
	})
	require.NoError(t, err)
	_, err = repo.Add(context.Background(), &models.ConferenceOrganizer{
		ConferenceID: conf.ID,
		UserID:       owner.ID,
		Role:         models.ConferenceOrganizerRoleOwner,
	})
	require.NoError(t, err)

	emails, err := repo.ListEmailsByConference(context.Background(), conf.ID)
	require.NoError(t, err)
	require.Len(t, emails, 2)
	assert.Equal(t, "owner@example.com", emails[0])
	assert.Equal(t, "admin@example.com", emails[1])
}
