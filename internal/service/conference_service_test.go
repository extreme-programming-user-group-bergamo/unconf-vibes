package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConferenceRepository struct {
	listFn      func(ctx context.Context) ([]*models.Conference, error)
	getBySlugFn func(ctx context.Context, slug string) (*models.Conference, error)
	createFn    func(ctx context.Context, conf *models.Conference) (*models.Conference, error)
}

func (m *mockConferenceRepository) List(ctx context.Context) ([]*models.Conference, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}

	return nil, nil
}

func (m *mockConferenceRepository) GetBySlug(ctx context.Context, slug string) (*models.Conference, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}

	return nil, nil
}

func (m *mockConferenceRepository) Create(ctx context.Context, conf *models.Conference) (*models.Conference, error) {
	if m.createFn != nil {
		return m.createFn(ctx, conf)
	}

	return nil, nil
}

func upcomingConference() *models.Conference {
	return &models.Conference{
		ID:          1,
		Slug:        "future-conf",
		Name:        "Future Conference",
		Description: "A future conference",
		Location:    "Future City",
		StartDate:   time.Now().Add(30 * 24 * time.Hour),
		EndDate:     time.Now().Add(33 * 24 * time.Hour),
		Capacity:    100,
		CreatedAt:   time.Now(),
	}
}

func activeConference() *models.Conference {
	return &models.Conference{
		ID:          2,
		Slug:        "active-conf",
		Name:        "Active Conference",
		Description: "An active conference",
		Location:    "Active City",
		StartDate:   time.Now().Add(-1 * 24 * time.Hour),
		EndDate:     time.Now().Add(2 * 24 * time.Hour),
		Capacity:    150,
		CreatedAt:   time.Now(),
	}
}

func pastConference() *models.Conference {
	return &models.Conference{
		ID:          3,
		Slug:        "past-conf",
		Name:        "Past Conference",
		Description: "A past conference",
		Location:    "Past City",
		StartDate:   time.Now().Add(-30 * 24 * time.Hour),
		EndDate:     time.Now().Add(-27 * 24 * time.Hour),
		Capacity:    200,
		CreatedAt:   time.Now(),
	}
}

func TestConferenceService_ListConferences_Success(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{upcomingConference(), pastConference()}, nil
		},
	}
	svc := NewConferenceService(repo)

	results, err := svc.ListConferences(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "upcoming", results[0].Status)
	assert.Equal(t, 0, results[0].AttendeeCount)
	assert.Equal(t, "past", results[1].Status)
	assert.Equal(t, 0, results[1].AttendeeCount)
}

func TestConferenceService_ListConferences_Empty(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{}, nil
		},
	}
	svc := NewConferenceService(repo)

	results, err := svc.ListConferences(context.Background())
	require.NoError(t, err)
	require.NotNil(t, results)
	assert.Empty(t, results)
}

func TestConferenceService_ListConferences_RepoError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return nil, repoErr
		},
	}
	svc := NewConferenceService(repo)

	_, err := svc.ListConferences(context.Background())
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestConferenceService_GetConference_Success(t *testing.T) {
	conf := activeConference()
	repo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, slug string) (*models.Conference, error) {
			assert.Equal(t, "active-conf", slug)
			return conf, nil
		},
	}
	svc := NewConferenceService(repo)

	result, err := svc.GetConference(context.Background(), "active-conf")
	require.NoError(t, err)
	assert.Equal(t, "active-conf", result.Slug)
	assert.Equal(t, "active", result.Status)
	assert.Equal(t, 0, result.AttendeeCount)
}

func TestConferenceService_GetConference_NotFound(t *testing.T) {
	repo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
			return nil, repository.ErrConferenceNotFound
		},
	}
	svc := NewConferenceService(repo)

	_, err := svc.GetConference(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}

func TestConferenceService_GetConference_RepoError(t *testing.T) {
	repoErr := errors.New("db timeout")
	repo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
			return nil, repoErr
		},
	}
	svc := NewConferenceService(repo)

	_, err := svc.GetConference(context.Background(), "some-slug")
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
}

func TestConferenceService_DeriveStatus_Upcoming(t *testing.T) {
	now := time.Now().UTC()
	conf := &models.Conference{
		StartDate: now.Add(30 * 24 * time.Hour),
		EndDate:   now.Add(33 * 24 * time.Hour),
	}

	status := DeriveStatusAt(conf, now)
	assert.Equal(t, "upcoming", status)
}

func TestConferenceService_DeriveStatus_Active(t *testing.T) {
	now := time.Now().UTC()
	conf := &models.Conference{
		StartDate: now.Add(-1 * 24 * time.Hour),
		EndDate:   now.Add(2 * 24 * time.Hour),
	}

	status := DeriveStatusAt(conf, now)
	assert.Equal(t, "active", status)
}

func TestConferenceService_DeriveStatus_Past(t *testing.T) {
	now := time.Now().UTC()
	conf := &models.Conference{
		StartDate: now.Add(-30 * 24 * time.Hour),
		EndDate:   now.Add(-27 * 24 * time.Hour),
	}

	status := DeriveStatusAt(conf, now)
	assert.Equal(t, "past", status)
}

func TestConferenceService_DeriveStatus_ActiveOnEndDate(t *testing.T) {
	now := time.Now().UTC()
	conf := &models.Conference{
		StartDate: now.Add(-2 * 24 * time.Hour),
		EndDate:   now, // end date is today — should still be "active"
	}

	status := DeriveStatusAt(conf, now)
	assert.Equal(t, "active", status)
}

func TestConferenceService_DeriveStatus_ActiveOnStartDate(t *testing.T) {
	now := time.Now().UTC()
	conf := &models.Conference{
		StartDate: now, // start date is today — should be "active"
		EndDate:   now.Add(3 * 24 * time.Hour),
	}

	status := DeriveStatusAt(conf, now)
	assert.Equal(t, "active", status)
}
