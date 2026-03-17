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

func newTestConference(slug string) *models.Conference {
	return &models.Conference{
		Slug:        slug,
		Name:        "Test Conference",
		Description: "A test conference",
		Location:    "Test City",
		StartDate:   time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
		Capacity:    200,
		HotelEmail:  "hotel@test.com",
	}
}

func TestConferenceRepository_Create_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	conf := newTestConference("test-conf")
	created, err := repo.Create(context.Background(), conf)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "test-conf", created.Slug)
	assert.Equal(t, "Test Conference", created.Name)
	assert.Equal(t, "A test conference", created.Description)
	assert.Equal(t, "Test City", created.Location)
	assert.Equal(t, 200, created.Capacity)
	assert.Equal(t, "hotel@test.com", created.HotelEmail)
	assert.False(t, created.CreatedAt.IsZero())
	assert.Equal(t, 2026, created.StartDate.Year())
	assert.Equal(t, time.October, created.StartDate.Month())
	assert.Equal(t, 7, created.StartDate.Day())
}

func TestConferenceRepository_Create_DuplicateSlug(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	_, err := repo.Create(context.Background(), newTestConference("dup-slug"))
	require.NoError(t, err)

	_, err = repo.Create(context.Background(), newTestConference("dup-slug"))
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrConferenceExists)
}

func TestConferenceRepository_GetBySlug_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	created, err := repo.Create(context.Background(), newTestConference("find-me"))
	require.NoError(t, err)

	fetched, err := repo.GetBySlug(context.Background(), "find-me")
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, "find-me", fetched.Slug)
	assert.Equal(t, "Test Conference", fetched.Name)
}

func TestConferenceRepository_GetBySlug_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	_, err := repo.GetBySlug(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, repository.ErrConferenceNotFound)
}

func TestConferenceRepository_List_ReturnsAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	conf1 := newTestConference("conf-earlier")
	conf1.StartDate = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	conf1.EndDate = time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC)

	conf2 := newTestConference("conf-later")
	conf2.Name = "Later Conference"
	conf2.StartDate = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	conf2.EndDate = time.Date(2026, 12, 4, 0, 0, 0, 0, time.UTC)

	_, err := repo.Create(context.Background(), conf1)
	require.NoError(t, err)
	_, err = repo.Create(context.Background(), conf2)
	require.NoError(t, err)

	list, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 2)
	// Ordered by start_date DESC — later conference first
	assert.Equal(t, "conf-later", list[0].Slug)
	assert.Equal(t, "conf-earlier", list[1].Slug)
}

func TestConferenceRepository_List_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConferenceRepository(db)

	list, err := repo.List(context.Background())
	require.NoError(t, err)
	require.NotNil(t, list)
	assert.Empty(t, list)
}
