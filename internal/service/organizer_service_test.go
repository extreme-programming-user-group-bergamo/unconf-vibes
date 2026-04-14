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

type mockOrganizerRepository struct {
	addFn                       func(ctx context.Context, organizer *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error)
	getByConferenceAndUserFn    func(ctx context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error)
	removeByConferenceAndUserFn func(ctx context.Context, conferenceID int64, userID int64) error
	isOrganizerFn               func(ctx context.Context, conferenceID int64, userID int64) (bool, error)
	isAnyOrganizerFn            func(ctx context.Context, userID int64) (bool, error)
}

func (m *mockOrganizerRepository) Add(ctx context.Context, organizer *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error) {
	if m.addFn != nil {
		return m.addFn(ctx, organizer)
	}
	return nil, nil
}

func (m *mockOrganizerRepository) GetByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error) {
	if m.getByConferenceAndUserFn != nil {
		return m.getByConferenceAndUserFn(ctx, conferenceID, userID)
	}
	return nil, repository.ErrConferenceOrganizerNotFound
}

func (m *mockOrganizerRepository) RemoveByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) error {
	if m.removeByConferenceAndUserFn != nil {
		return m.removeByConferenceAndUserFn(ctx, conferenceID, userID)
	}
	return nil
}

func (m *mockOrganizerRepository) IsOrganizer(ctx context.Context, conferenceID int64, userID int64) (bool, error) {
	if m.isOrganizerFn != nil {
		return m.isOrganizerFn(ctx, conferenceID, userID)
	}
	return false, nil
}

func (m *mockOrganizerRepository) IsOrganizerForAnyConference(ctx context.Context, userID int64) (bool, error) {
	if m.isAnyOrganizerFn != nil {
		return m.isAnyOrganizerFn(ctx, userID)
	}
	return false, nil
}

type mockSimpleUserRepository struct {
	getByIDFn func(ctx context.Context, id int64) (*models.User, error)
}

func (m *mockSimpleUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSimpleUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, repository.ErrUserNotFound
}
func (m *mockSimpleUserRepository) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockSimpleUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func conferenceForOrganizerTests() *models.Conference {
	return &models.Conference{
		ID:        99,
		Slug:      "organizer-conf",
		Name:      "Organizer Conf",
		Location:  "Town",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		Capacity:  100,
	}
}

func TestOrganizerService_AddOrganizer_OwnerCanAdd(t *testing.T) {
	confRepo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, slug string) (*models.Conference, error) {
			assert.Equal(t, "organizer-conf", slug)
			return conferenceForOrganizerTests(), nil
		},
	}
	organizerRepo := &mockOrganizerRepository{
		getByConferenceAndUserFn: func(_ context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error) {
			if userID == 1 {
				return &models.ConferenceOrganizer{ConferenceID: conferenceID, UserID: userID, Role: models.ConferenceOrganizerRoleOwner}, nil
			}
			return nil, repository.ErrConferenceOrganizerNotFound
		},
		addFn: func(_ context.Context, organizer *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error) {
			return &models.ConferenceOrganizer{
				ID:           5,
				ConferenceID: organizer.ConferenceID,
				UserID:       organizer.UserID,
				Role:         organizer.Role,
			}, nil
		},
	}
	userRepo := &mockSimpleUserRepository{
		getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id}, nil
		},
	}

	svc := NewOrganizerService(confRepo, organizerRepo, userRepo)
	added, err := svc.AddOrganizer(context.Background(), "organizer-conf", 1, 2, "admin")
	require.NoError(t, err)
	require.NotNil(t, added)
	assert.Equal(t, models.ConferenceOrganizerRoleAdmin, added.Role)
}

func TestOrganizerService_AddOrganizer_AdminCannotAdd(t *testing.T) {
	confRepo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
			return conferenceForOrganizerTests(), nil
		},
	}
	organizerRepo := &mockOrganizerRepository{
		getByConferenceAndUserFn: func(_ context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error) {
			return &models.ConferenceOrganizer{ConferenceID: conferenceID, UserID: userID, Role: models.ConferenceOrganizerRoleAdmin}, nil
		},
	}
	userRepo := &mockSimpleUserRepository{
		getByIDFn: func(_ context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id}, nil
		},
	}

	svc := NewOrganizerService(confRepo, organizerRepo, userRepo)
	_, err := svc.AddOrganizer(context.Background(), "organizer-conf", 1, 2, "admin")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrOwnerRequired)
}

func TestOrganizerService_IsOrganizerForConference_NonMember(t *testing.T) {
	confRepo := &mockConferenceRepository{
		getBySlugFn: func(_ context.Context, _ string) (*models.Conference, error) {
			return conferenceForOrganizerTests(), nil
		},
	}
	organizerRepo := &mockOrganizerRepository{
		isOrganizerFn: func(_ context.Context, _ int64, _ int64) (bool, error) {
			return false, nil
		},
	}

	svc := NewOrganizerService(confRepo, organizerRepo, &mockSimpleUserRepository{})
	ok, err := svc.IsOrganizerForConference(context.Background(), "organizer-conf", 12)
	require.NoError(t, err)
	assert.False(t, ok)
}
