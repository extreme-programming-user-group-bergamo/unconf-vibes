package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

var _ repository.EmailLogRepository = (*EmailLogRepository)(nil)

type EmailLogRepository struct {
	db *sql.DB
}

func NewEmailLogRepository(db *sql.DB) *EmailLogRepository {
	return &EmailLogRepository{db: db}
}

func (r *EmailLogRepository) Create(ctx context.Context, log *models.EmailLog) (*models.EmailLog, error) {
	if log == nil {
		return nil, fmt.Errorf("failed to create email log: log is nil")
	}

	query := `
		INSERT INTO email_logs (booking_id, email_type, recipient, subject, status, attempt, error_details)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, booking_id, email_type, recipient, subject, status, attempt, COALESCE(error_details, ''), created_at
	`

	created, err := scanEmailLog(r.db.QueryRowContext(ctx, query,
		log.BookingID,
		log.EmailType,
		log.Recipient,
		log.Subject,
		log.Status,
		log.Attempt,
		log.ErrorDetails,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create email log: %w", err)
	}

	return created, nil
}

func (r *EmailLogRepository) ListByBookingID(ctx context.Context, bookingID int64) ([]*models.EmailLog, error) {
	query := `
		SELECT id, booking_id, email_type, recipient, subject, status, attempt, COALESCE(error_details, ''), created_at
		FROM email_logs
		WHERE booking_id = ?
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to list email logs by booking id: %w", err)
	}
	defer func() { _ = rows.Close() }()

	logs := make([]*models.EmailLog, 0)
	for rows.Next() {
		log, scanErr := scanEmailLog(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan email log row: %w", scanErr)
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate email log rows: %w", err)
	}

	if len(logs) == 0 {
		return nil, repository.ErrEmailLogNotFound
	}

	return logs, nil
}

func scanEmailLog(s scanner) (*models.EmailLog, error) {
	var log models.EmailLog
	var status string

	err := s.Scan(
		&log.ID,
		&log.BookingID,
		&log.EmailType,
		&log.Recipient,
		&log.Subject,
		&status,
		&log.Attempt,
		&log.ErrorDetails,
		&log.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrEmailLogNotFound
		}
		return nil, err
	}

	log.Status = models.EmailLogStatus(status)
	return &log, nil
}
