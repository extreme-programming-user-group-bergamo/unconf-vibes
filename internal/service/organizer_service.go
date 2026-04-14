package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

type OrganizerService struct {
	confRepo      repository.ConferenceRepository
	organizerRepo repository.ConferenceOrganizerRepository
	userRepo      repository.UserRepository
}

func NewOrganizerService(
	confRepo repository.ConferenceRepository,
	organizerRepo repository.ConferenceOrganizerRepository,
	userRepo repository.UserRepository,
) *OrganizerService {
	return &OrganizerService{
		confRepo:      confRepo,
		organizerRepo: organizerRepo,
		userRepo:      userRepo,
	}
}

func (s *OrganizerService) IsOrganizerForConference(ctx context.Context, slug string, userID int64) (bool, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return false, fmt.Errorf("failed to check organizer role: %w", ErrConferenceNotFound)
		}
		return false, fmt.Errorf("failed to check organizer role: %w", err)
	}

	isOrganizer, err := s.organizerRepo.IsOrganizer(ctx, conf.ID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check organizer role: %w", err)
	}

	return isOrganizer, nil
}

func (s *OrganizerService) IsOwnerForConference(ctx context.Context, slug string, userID int64) (bool, error) {
	conf, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return false, fmt.Errorf("failed to check organizer owner role: %w", ErrConferenceNotFound)
		}
		return false, fmt.Errorf("failed to check organizer owner role: %w", err)
	}

	organizer, err := s.organizerRepo.GetByConferenceAndUser(ctx, conf.ID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceOrganizerNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check organizer owner role: %w", err)
	}

	return organizer.Role == models.ConferenceOrganizerRoleOwner, nil
}

func (s *OrganizerService) AddOrganizer(
	ctx context.Context,
	slug string,
	requesterUserID int64,
	targetUserID int64,
	role string,
) (*models.ConferenceOrganizer, error) {
	conference, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return nil, fmt.Errorf("failed to add organizer: %w", ErrConferenceNotFound)
		}
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	if err := s.requireOwner(ctx, conference.ID, requesterUserID); err != nil {
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	if _, err := s.userRepo.GetByID(ctx, targetUserID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to add organizer: %w", ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	normalizedRole := models.ConferenceOrganizerRole(strings.ToLower(strings.TrimSpace(role)))
	if normalizedRole != models.ConferenceOrganizerRoleOwner && normalizedRole != models.ConferenceOrganizerRoleAdmin {
		return nil, fmt.Errorf("failed to add organizer: invalid role")
	}

	created, err := s.organizerRepo.Add(ctx, &models.ConferenceOrganizer{
		ConferenceID: conference.ID,
		UserID:       targetUserID,
		Role:         normalizedRole,
	})
	if err != nil {
		if errors.Is(err, repository.ErrConferenceOrganizerExists) {
			return nil, fmt.Errorf("failed to add organizer: %w", ErrOrganizerExists)
		}
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	return created, nil
}

func (s *OrganizerService) RemoveOrganizer(ctx context.Context, slug string, requesterUserID int64, targetUserID int64) error {
	conference, err := s.confRepo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceNotFound) {
			return fmt.Errorf("failed to remove organizer: %w", ErrConferenceNotFound)
		}
		return fmt.Errorf("failed to remove organizer: %w", err)
	}

	if err := s.requireOwner(ctx, conference.ID, requesterUserID); err != nil {
		return fmt.Errorf("failed to remove organizer: %w", err)
	}
	if requesterUserID == targetUserID {
		return fmt.Errorf("failed to remove organizer: %w", ErrOwnerRequired)
	}

	targetMembership, err := s.organizerRepo.GetByConferenceAndUser(ctx, conference.ID, targetUserID)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceOrganizerNotFound) {
			return fmt.Errorf("failed to remove organizer: %w", ErrOrganizerNotFound)
		}
		return fmt.Errorf("failed to remove organizer: %w", err)
	}
	if targetMembership.Role == models.ConferenceOrganizerRoleOwner {
		return fmt.Errorf("failed to remove organizer: %w", ErrOwnerRequired)
	}

	if err := s.organizerRepo.RemoveByConferenceAndUser(ctx, conference.ID, targetUserID); err != nil {
		if errors.Is(err, repository.ErrConferenceOrganizerNotFound) {
			return fmt.Errorf("failed to remove organizer: %w", ErrOrganizerNotFound)
		}
		return fmt.Errorf("failed to remove organizer: %w", err)
	}

	return nil
}

func (s *OrganizerService) requireOwner(ctx context.Context, conferenceID int64, userID int64) error {
	membership, err := s.organizerRepo.GetByConferenceAndUser(ctx, conferenceID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrConferenceOrganizerNotFound) {
			return ErrOrganizerForbidden
		}
		return err
	}
	if membership.Role != models.ConferenceOrganizerRoleOwner {
		return ErrOwnerRequired
	}

	return nil
}
