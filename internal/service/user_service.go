package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

var validPrivacySettings = map[string]bool{
	"public":           true,
	"private":          true,
	"connections_only": true,
}

// UpdateProfileInput holds optional fields for updating a user profile.
type UpdateProfileInput struct {
	DisplayName    *string
	PrivacySetting *string
}

// UserService handles user profile business logic.
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService with the given repository.
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile retrieves a user's profile by ID.
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to get user profile: %w", ErrUserNotFound)
		}

		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user, nil
}

// UpdateProfile updates mutable profile fields for the given user.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, input UpdateProfileInput) (*models.User, error) {
	if input.PrivacySetting != nil && !validPrivacySettings[*input.PrivacySetting] {
		return nil, fmt.Errorf("failed to update user profile: %w", ErrInvalidPrivacySetting)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to update user profile: %w", ErrUserNotFound)
		}

		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}

	if input.PrivacySetting != nil {
		user.PrivacySetting = *input.PrivacySetting
	}

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return updated, nil
}
