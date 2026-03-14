package service

import (
	"context"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestUser() *models.User {
	return &models.User{
		ID:             1,
		GitHubID:       "12345",
		Email:          "test@example.com",
		DisplayName:    "Test User",
		PrivacySetting: "public",
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

func TestUserService_GetProfile_Success(t *testing.T) {
	user := newTestUser()
	userRepo := &repository.MockUserRepository{
		GetByIDFunc: func(_ context.Context, id int64) (*models.User, error) {
			assert.Equal(t, int64(1), id)
			return user, nil
		},
	}

	svc := NewUserService(userRepo)
	result, err := svc.GetProfile(context.Background(), 1)

	require.NoError(t, err)
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.DisplayName, result.DisplayName)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	userRepo := &repository.MockUserRepository{
		GetByIDFunc: func(_ context.Context, _ int64) (*models.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	svc := NewUserService(userRepo)
	_, err := svc.GetProfile(context.Background(), 999)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserService_UpdateProfile_DisplayName(t *testing.T) {
	user := newTestUser()
	updatedUser := *user
	updatedUser.DisplayName = "New Name"

	userRepo := &repository.MockUserRepository{
		GetByIDFunc: func(_ context.Context, id int64) (*models.User, error) {
			return user, nil
		},
		UpdateFunc: func(_ context.Context, u *models.User) (*models.User, error) {
			assert.Equal(t, "New Name", u.DisplayName)
			return &updatedUser, nil
		},
	}

	svc := NewUserService(userRepo)
	name := "New Name"
	result, err := svc.UpdateProfile(context.Background(), 1, UpdateProfileInput{DisplayName: &name})

	require.NoError(t, err)
	assert.Equal(t, "New Name", result.DisplayName)
}

func TestUserService_UpdateProfile_PrivacySetting(t *testing.T) {
	user := newTestUser()
	updatedUser := *user
	updatedUser.PrivacySetting = "private"

	userRepo := &repository.MockUserRepository{
		GetByIDFunc: func(_ context.Context, _ int64) (*models.User, error) {
			return user, nil
		},
		UpdateFunc: func(_ context.Context, u *models.User) (*models.User, error) {
			assert.Equal(t, "private", u.PrivacySetting)
			return &updatedUser, nil
		},
	}

	svc := NewUserService(userRepo)
	setting := "private"
	result, err := svc.UpdateProfile(context.Background(), 1, UpdateProfileInput{PrivacySetting: &setting})

	require.NoError(t, err)
	assert.Equal(t, "private", result.PrivacySetting)
}

func TestUserService_UpdateProfile_InvalidPrivacySetting(t *testing.T) {
	userRepo := &repository.MockUserRepository{}
	svc := NewUserService(userRepo)

	setting := "invalid_value"
	_, err := svc.UpdateProfile(context.Background(), 1, UpdateProfileInput{PrivacySetting: &setting})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPrivacySetting)
}

func TestUserService_UpdateProfile_NotFound(t *testing.T) {
	userRepo := &repository.MockUserRepository{
		GetByIDFunc: func(_ context.Context, _ int64) (*models.User, error) {
			return nil, repository.ErrUserNotFound
		},
	}

	svc := NewUserService(userRepo)
	name := "New Name"
	_, err := svc.UpdateProfile(context.Background(), 999, UpdateProfileInput{DisplayName: &name})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
}
