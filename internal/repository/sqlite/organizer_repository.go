package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

var _ repository.ConferenceOrganizerRepository = (*OrganizerRepository)(nil)

type OrganizerRepository struct {
	db *sql.DB
}

func NewOrganizerRepository(db *sql.DB) *OrganizerRepository {
	return &OrganizerRepository{db: db}
}

func (r *OrganizerRepository) Add(ctx context.Context, organizer *models.ConferenceOrganizer) (*models.ConferenceOrganizer, error) {
	if organizer == nil {
		return nil, fmt.Errorf("failed to add organizer: organizer is nil")
	}

	query := `
		INSERT INTO conference_organizers (conference_id, user_id, role)
		VALUES (?, ?, ?)
		RETURNING id, conference_id, user_id, role, created_at
	`

	created, err := scanConferenceOrganizer(r.db.QueryRowContext(ctx, query, organizer.ConferenceID, organizer.UserID, organizer.Role))
	if err != nil {
		if isConferenceOrganizerUniqueConstraintError(err) {
			return nil, repository.ErrConferenceOrganizerExists
		}
		return nil, fmt.Errorf("failed to add organizer: %w", err)
	}

	return created, nil
}

func (r *OrganizerRepository) GetByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) (*models.ConferenceOrganizer, error) {
	query := `
		SELECT id, conference_id, user_id, role, created_at
		FROM conference_organizers
		WHERE conference_id = ? AND user_id = ?
	`

	organizer, err := scanConferenceOrganizer(r.db.QueryRowContext(ctx, query, conferenceID, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrConferenceOrganizerNotFound
		}
		return nil, fmt.Errorf("failed to fetch organizer membership: %w", err)
	}

	return organizer, nil
}

func (r *OrganizerRepository) RemoveByConferenceAndUser(ctx context.Context, conferenceID int64, userID int64) error {
	query := `
		DELETE FROM conference_organizers
		WHERE conference_id = ? AND user_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, conferenceID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove organizer membership: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to remove organizer membership: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrConferenceOrganizerNotFound
	}

	return nil
}

func (r *OrganizerRepository) IsOrganizer(ctx context.Context, conferenceID int64, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM conference_organizers
			WHERE conference_id = ? AND user_id = ?
		)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, conferenceID, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check organizer membership: %w", err)
	}

	return exists, nil
}

func scanConferenceOrganizer(s scanner) (*models.ConferenceOrganizer, error) {
	var organizer models.ConferenceOrganizer
	err := s.Scan(&organizer.ID, &organizer.ConferenceID, &organizer.UserID, &organizer.Role, &organizer.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &organizer, nil
}

func isConferenceOrganizerUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: conference_organizers.conference_id, conference_organizers.user_id")
}
