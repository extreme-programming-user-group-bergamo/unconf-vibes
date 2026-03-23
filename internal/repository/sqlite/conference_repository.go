package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

type ConferenceRepository struct {
	db *sql.DB
}

func NewConferenceRepository(db *sql.DB) *ConferenceRepository {
	return &ConferenceRepository{db: db}
}

func (r *ConferenceRepository) Create(ctx context.Context, conf *models.Conference) (*models.Conference, error) {
	if conf == nil {
		return nil, fmt.Errorf("failed to create conference: conference is nil")
	}

	query := `
		INSERT INTO conferences (slug, name, description, location, start_date, end_date, capacity, hotel_email)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, slug, name, description, location, start_date, end_date, capacity, hotel_email, created_at
	`

	created, err := scanConference(r.db.QueryRowContext(ctx, query,
		conf.Slug, conf.Name, conf.Description, conf.Location,
		conf.StartDate, conf.EndDate,
		conf.Capacity, conf.HotelEmail,
	))
	if err != nil {
		if isConferenceUniqueConstraintError(err) {
			return nil, repository.ErrConferenceExists
		}

		return nil, fmt.Errorf("failed to create conference: %w", err)
	}

	slog.Debug("conference created", "conference_id", created.ID, "slug", created.Slug)
	return created, nil
}

func (r *ConferenceRepository) GetBySlug(ctx context.Context, slug string) (*models.Conference, error) {
	query := `
		SELECT id, slug, name, description, location, start_date, end_date, capacity, hotel_email, created_at
		FROM conferences
		WHERE slug = ?
	`

	conf, err := scanConference(r.db.QueryRowContext(ctx, query, slug))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrConferenceNotFound
		}

		return nil, fmt.Errorf("failed to fetch conference by slug: %w", err)
	}

	return conf, nil
}

func (r *ConferenceRepository) List(ctx context.Context) ([]*models.Conference, error) {
	query := `
		SELECT id, slug, name, description, location, start_date, end_date, capacity, hotel_email, created_at
		FROM conferences
		ORDER BY start_date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list conferences: %w", err)
	}
	defer func() { _ = rows.Close() }()

	conferences := make([]*models.Conference, 0)
	for rows.Next() {
		conf, err := scanConference(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conference row: %w", err)
		}

		conferences = append(conferences, conf)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate conference rows: %w", err)
	}

	slog.Debug("conferences listed", "count", len(conferences))
	return conferences, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanConference(s scanner) (*models.Conference, error) {
	var conf models.Conference
	var description sql.NullString
	var hotelEmail sql.NullString

	err := s.Scan(
		&conf.ID,
		&conf.Slug,
		&conf.Name,
		&description,
		&conf.Location,
		&conf.StartDate,
		&conf.EndDate,
		&conf.Capacity,
		&hotelEmail,
		&conf.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	conf.Description = description.String
	conf.HotelEmail = hotelEmail.String

	return &conf, nil
}

func isConferenceUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: conferences.slug")
}

var _ repository.ConferenceRepository = (*ConferenceRepository)(nil)
