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
	listFn            func(ctx context.Context) ([]*models.Conference, error)
	getBySlugFn       func(ctx context.Context, slug string) (*models.Conference, error)
	createFn          func(ctx context.Context, conf *models.Conference) (*models.Conference, error)
	createWithOwnerFn func(ctx context.Context, conf *models.Conference, ownerUserID int64) (*models.Conference, error)
	updateBySlugFn    func(ctx context.Context, slug string, conf *models.Conference) (*models.Conference, error)
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

func (m *mockConferenceRepository) CreateWithOwner(ctx context.Context, conf *models.Conference, ownerUserID int64) (*models.Conference, error) {
	if m.createWithOwnerFn != nil {
		return m.createWithOwnerFn(ctx, conf, ownerUserID)
	}

	return nil, nil
}

func (m *mockConferenceRepository) UpdateBySlug(ctx context.Context, slug string, conf *models.Conference) (*models.Conference, error) {
	if m.updateBySlugFn != nil {
		return m.updateBySlugFn(ctx, slug, conf)
	}

	return nil, nil
}

type mockConferenceOrganizerRepository struct {
	isOrganizerForAnyConferenceFn func(ctx context.Context, userID int64) (bool, error)
}

func (m *mockConferenceOrganizerRepository) Add(context.Context, *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error) {
	panic("not implemented")
}

func (m *mockConferenceOrganizerRepository) GetByConferenceAndUser(context.Context, int64, int64) (*models.ConferenceOrganizer, error) {
	panic("not implemented")
}

func (m *mockConferenceOrganizerRepository) ListEmailsByConference(context.Context, int64) ([]string, error) {
	return []string{}, nil
}

func (m *mockConferenceOrganizerRepository) RemoveByConferenceAndUser(context.Context, int64, int64) error {
	panic("not implemented")
}

func (m *mockConferenceOrganizerRepository) IsOrganizer(context.Context, int64, int64) (bool, error) {
	panic("not implemented")
}

func (m *mockConferenceOrganizerRepository) IsOrganizerForAnyConference(ctx context.Context, userID int64) (bool, error) {
	if m.isOrganizerForAnyConferenceFn != nil {
		return m.isOrganizerForAnyConferenceFn(ctx, userID)
	}

	return true, nil
}

func newMockConferenceOrganizerRepository(isOrganizer bool) *mockConferenceOrganizerRepository {
	return &mockConferenceOrganizerRepository{
		isOrganizerForAnyConferenceFn: func(context.Context, int64) (bool, error) {
			return isOrganizer, nil
		},
	}
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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

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

func TestConferenceService_ListConferences_AttendeeCountFromBookings(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{upcomingConference()}, nil
		},
	}
	bookingRepo := &mockBookingRepository{
		countByConferenceFn: func(_ context.Context, confID int64) (int, error) {
			assert.Equal(t, int64(1), confID)
			return 5, nil
		},
	}
	svc := NewConferenceService(repo, bookingRepo, newMockConferenceOrganizerRepository(true))

	results, err := svc.ListConferences(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 5, results[0].AttendeeCount)
}

func TestConferenceService_GetConference_AttendeeCountFromBookings(t *testing.T) {
	conf := activeConference()
	repo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
			return conf, nil
		},
	}
	bookingRepo := &mockBookingRepository{
		countByConferenceFn: func(_ context.Context, _ int64) (int, error) {
			return 3, nil
		},
	}
	svc := NewConferenceService(repo, bookingRepo, newMockConferenceOrganizerRepository(true))

	result, err := svc.GetConference(context.Background(), "active-conf")
	require.NoError(t, err)
	assert.Equal(t, 3, result.AttendeeCount)
}

func TestConferenceService_ListConferences_BookingCountErrorDefaultsToZero(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{upcomingConference()}, nil
		},
	}
	bookingRepo := &mockBookingRepository{
		countByConferenceFn: func(_ context.Context, _ int64) (int, error) {
			return 0, errors.New("booking count failure")
		},
	}
	svc := NewConferenceService(repo, bookingRepo, newMockConferenceOrganizerRepository(true))

	results, err := svc.ListConferences(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 0, results[0].AttendeeCount)
}

func TestConferenceService_CreateConference_Success(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{{ID: 1, Slug: "existing"}}, nil
		},
		createWithOwnerFn: func(_ context.Context, conf *models.Conference, ownerUserID int64) (*models.Conference, error) {
			assert.Equal(t, int64(77), ownerUserID)
			assert.Equal(t, "new-conf", conf.Slug)
			return &models.Conference{
				ID:          123,
				Slug:        conf.Slug,
				Name:        conf.Name,
				Description: conf.Description,
				Location:    conf.Location,
				StartDate:   conf.StartDate,
				EndDate:     conf.EndDate,
				Capacity:    conf.Capacity,
				HotelEmail:  conf.HotelEmail,
				CreatedAt:   time.Now(),
			}, nil
		},
	}

	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))
	result, err := svc.CreateConference(context.Background(), 77, CreateConferenceInput{
		Slug:        "new-conf",
		Name:        "New Conf",
		Description: "desc",
		Location:    "Berlin",
		StartDate:   time.Now().Add(24 * time.Hour),
		EndDate:     time.Now().Add(48 * time.Hour),
		Capacity:    100,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "new-conf", result.Slug)
}

func TestConferenceService_CreateConference_DuplicateSlug(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{{ID: 1, Slug: "existing"}}, nil
		},
		createWithOwnerFn: func(_ context.Context, _ *models.Conference, _ int64) (*models.Conference, error) {
			return nil, repository.ErrConferenceExists
		},
	}

	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))
	_, err := svc.CreateConference(context.Background(), 55, CreateConferenceInput{
		Slug:      "dup",
		Name:      "Duplicate",
		Location:  "Paris",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  100,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceExists)
}

func TestConferenceService_CreateConference_OrganizerRequired(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{{ID: 1, Slug: "existing"}}, nil
		},
	}
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(false))

	_, err := svc.CreateConference(context.Background(), 55, CreateConferenceInput{
		Slug:      "restricted",
		Name:      "Restricted",
		Location:  "Paris",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  100,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOrganizerForbidden)
}

func TestConferenceService_CreateConference_AllowsBootstrapWhenNoConferencesExist(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{}, nil
		},
		createWithOwnerFn: func(_ context.Context, conf *models.Conference, ownerUserID int64) (*models.Conference, error) {
			assert.Equal(t, int64(99), ownerUserID)
			assert.Equal(t, "bootstrap-conf", conf.Slug)
			return &models.Conference{
				ID:        321,
				Slug:      conf.Slug,
				Name:      conf.Name,
				Location:  conf.Location,
				StartDate: conf.StartDate,
				EndDate:   conf.EndDate,
				Capacity:  conf.Capacity,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(false))

	result, err := svc.CreateConference(context.Background(), 99, CreateConferenceInput{
		Slug:      "bootstrap-conf",
		Name:      "Bootstrap",
		Location:  "Berlin",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  50,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "bootstrap-conf", result.Slug)
}

func TestConferenceService_CreateConference_NormalizesSlug(t *testing.T) {
	repo := &mockConferenceRepository{
		createWithOwnerFn: func(_ context.Context, conf *models.Conference, _ int64) (*models.Conference, error) {
			assert.Equal(t, "new-conf-2026", conf.Slug)
			return &models.Conference{
				ID:        9,
				Slug:      conf.Slug,
				Name:      conf.Name,
				Location:  conf.Location,
				StartDate: conf.StartDate,
				EndDate:   conf.EndDate,
				Capacity:  conf.Capacity,
				CreatedAt: time.Now(),
			}, nil
		},
	}
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(false))

	created, err := svc.CreateConference(context.Background(), 44, CreateConferenceInput{
		Slug:      "  New_Conf 2026  ",
		Name:      "New Conf",
		Location:  "Berlin",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  20,
	})
	require.NoError(t, err)
	assert.Equal(t, "new-conf-2026", created.Slug)
}

func TestConferenceService_CreateConference_InvalidSlugRejected(t *testing.T) {
	repo := &mockConferenceRepository{
		listFn: func(_ context.Context) ([]*models.Conference, error) {
			return []*models.Conference{}, nil
		},
	}
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(false))

	_, err := svc.CreateConference(context.Background(), 44, CreateConferenceInput{
		Slug:      "bad.slug",
		Name:      "Bad Slug",
		Location:  "Berlin",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  20,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidConferenceInput)
}

func TestConferenceService_UpdateConference_Success(t *testing.T) {
	repo := &mockConferenceRepository{
		updateBySlugFn: func(_ context.Context, slug string, conf *models.Conference) (*models.Conference, error) {
			assert.Equal(t, "socrates-26", slug)
			return &models.Conference{
				ID:          77,
				Slug:        "socrates-26",
				Name:        conf.Name,
				Description: conf.Description,
				Location:    conf.Location,
				StartDate:   conf.StartDate,
				EndDate:     conf.EndDate,
				Capacity:    conf.Capacity,
				CreatedAt:   time.Now(),
			}, nil
		},
	}
	bookingRepo := &mockBookingRepository{
		countByConferenceFn: func(_ context.Context, conferenceID int64) (int, error) {
			assert.Equal(t, int64(77), conferenceID)
			return 8, nil
		},
	}
	svc := NewConferenceService(repo, bookingRepo, newMockConferenceOrganizerRepository(true))

	updated, err := svc.UpdateConference(context.Background(), "socrates-26", UpdateConferenceInput{
		Name:        "SoCraTes 2026 Updated",
		Description: "updated",
		Location:    "Berlin",
		StartDate:   time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC),
		Capacity:    250,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "socrates-26", updated.Slug)
	assert.Equal(t, 8, updated.AttendeeCount)
}

func TestConferenceService_UpdateConference_NotFound(t *testing.T) {
	repo := &mockConferenceRepository{
		updateBySlugFn: func(_ context.Context, _ string, _ *models.Conference) (*models.Conference, error) {
			return nil, repository.ErrConferenceNotFound
		},
	}
	svc := NewConferenceService(repo, &mockBookingRepository{}, newMockConferenceOrganizerRepository(true))

	_, err := svc.UpdateConference(context.Background(), "missing", UpdateConferenceInput{
		Name:      "Missing",
		Location:  "Nowhere",
		StartDate: time.Now().Add(24 * time.Hour),
		EndDate:   time.Now().Add(48 * time.Hour),
		Capacity:  10,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrConferenceNotFound)
}
