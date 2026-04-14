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

var _ repository.RoommateRequestRepository = (*RoommateRequestRepository)(nil)

type RoommateRequestRepository struct {
	db *sql.DB
}

func NewRoommateRequestRepository(db *sql.DB) *RoommateRequestRepository {
	return &RoommateRequestRepository{db: db}
}

func (r *RoommateRequestRepository) Create(ctx context.Context, request *models.RoommateRequest) (*models.RoommateRequest, error) {
	if request == nil {
		return nil, fmt.Errorf("failed to create roommate request: request is nil")
	}

	query := `
		INSERT INTO roommate_requests (requester_id, target_id, room_id, status)
		VALUES (?, ?, ?, ?)
		RETURNING id, requester_id, target_id, room_id, status, created_at
	`

	created, err := scanRoommateRequest(r.db.QueryRowContext(ctx, query,
		request.RequesterID,
		request.TargetID,
		request.RoomID,
		request.Status,
	))
	if err != nil {
		if isRoommateRequestUniqueConstraintError(err) {
			return nil, repository.ErrRoommateRequestExists
		}

		return nil, fmt.Errorf("failed to create roommate request: %w", err)
	}

	slog.Debug("roommate request created", "request_id", created.ID, "requester_id", created.RequesterID, "target_id", created.TargetID)
	return created, nil
}

func (r *RoommateRequestRepository) GetByID(ctx context.Context, id int64) (*models.RoommateRequest, error) {
	query := `
		SELECT id, requester_id, target_id, room_id, status, created_at
		FROM roommate_requests
		WHERE id = ?
	`

	result, err := scanRoommateRequest(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRoommateRequestNotFound
		}
		return nil, fmt.Errorf("failed to fetch roommate request by id: %w", err)
	}

	return result, nil
}

func (r *RoommateRequestRepository) ListByUser(ctx context.Context, userID int64) ([]*models.RoommateRequest, error) {
	query := `
		SELECT id, requester_id, target_id, room_id, status, created_at
		FROM roommate_requests
		WHERE requester_id = ? OR target_id = ?
		ORDER BY created_at DESC, id DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roommate requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	requests := make([]*models.RoommateRequest, 0)
	for rows.Next() {
		request, scanErr := scanRoommateRequest(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan roommate request row: %w", scanErr)
		}
		requests = append(requests, request)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate roommate request rows: %w", err)
	}

	return requests, nil
}

func (r *RoommateRequestRepository) UpdateStatus(ctx context.Context, id int64, status models.RoommateRequestStatus) (*models.RoommateRequest, error) {
	query := `
		UPDATE roommate_requests
		SET status = ?
		WHERE id = ?
		RETURNING id, requester_id, target_id, room_id, status, created_at
	`

	updated, err := scanRoommateRequest(r.db.QueryRowContext(ctx, query, status, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRoommateRequestNotFound
		}
		return nil, fmt.Errorf("failed to update roommate request status: %w", err)
	}

	return updated, nil
}

func (r *RoommateRequestRepository) CancelPendingOutgoingByConference(ctx context.Context, requesterID, conferenceID int64) (int64, error) {
	query := `
		UPDATE roommate_requests
		SET status = 'cancelled'
		WHERE requester_id = ?
		  AND status = 'pending'
		  AND room_id IN (
		    SELECT id FROM rooms WHERE conference_id = ?
		  )
	`

	result, err := r.db.ExecContext(ctx, query, requesterID, conferenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel pending outgoing roommate requests by conference: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to read cancelled roommate request count: %w", err)
	}

	return rowsAffected, nil
}

func scanRoommateRequest(s scanner) (*models.RoommateRequest, error) {
	var req models.RoommateRequest
	if err := s.Scan(
		&req.ID,
		&req.RequesterID,
		&req.TargetID,
		&req.RoomID,
		&req.Status,
		&req.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &req, nil
}

func isRoommateRequestUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: roommate_requests.requester_id, roommate_requests.target_id, roommate_requests.room_id")
}
