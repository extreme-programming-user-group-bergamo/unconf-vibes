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

var _ repository.BookingRepository = (*BookingRepository)(nil)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	if booking == nil {
		return nil, fmt.Errorf("failed to create booking: booking is nil")
	}

	query := `
		INSERT INTO bookings (room_id, user_id, conference_id, status, privacy_setting, notes)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
	`

	created, err := scanBooking(r.db.QueryRowContext(ctx, query,
		booking.RoomID, booking.UserID, booking.ConferenceID,
		booking.Status, booking.PrivacySetting, booking.Notes,
	))
	if err != nil {
		if isBookingUniqueConstraintError(err) {
			return nil, repository.ErrBookingExists
		}

		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	slog.Debug("booking created", "booking_id", created.ID, "user_id", created.UserID, "room_id", created.RoomID)
	return created, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id int64) (*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE id = ?
	`

	booking, err := scanBooking(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}

		return nil, fmt.Errorf("failed to fetch booking by id: %w", err)
	}

	return booking, nil
}

func (r *BookingRepository) ListByRoom(ctx context.Context, roomID int64) ([]*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE room_id = ? AND status != 'cancelled'
		ORDER BY created_at
	`

	return r.listBookings(ctx, query, roomID)
}

func (r *BookingRepository) ListByConference(ctx context.Context, conferenceID int64) ([]*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE conference_id = ? AND status != 'cancelled'
		ORDER BY created_at
	`

	return r.listBookings(ctx, query, conferenceID)
}

func (r *BookingRepository) ListByConferenceIncludingCancelled(ctx context.Context, conferenceID int64) ([]*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE conference_id = ?
		ORDER BY created_at
	`

	return r.listBookings(ctx, query, conferenceID)
}

func (r *BookingRepository) ListByUser(ctx context.Context, userID int64) ([]*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE user_id = ? AND status != 'cancelled'
		ORDER BY created_at DESC
	`

	return r.listBookings(ctx, query, userID)
}

func (r *BookingRepository) CountByConference(ctx context.Context, conferenceID int64) (int, error) {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM bookings
		WHERE conference_id = ? AND status != 'cancelled'
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, conferenceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookings by conference: %w", err)
	}

	slog.Debug("bookings counted by conference", "conference_id", conferenceID, "count", count)
	return count, nil
}

func (r *BookingRepository) GetActiveByUserAndConference(ctx context.Context, userID, conferenceID int64) (*models.Booking, error) {
	query := `
		SELECT id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
		FROM bookings
		WHERE user_id = ? AND conference_id = ? AND status != 'cancelled'
		ORDER BY created_at DESC
		LIMIT 1
	`

	booking, err := scanBooking(r.db.QueryRowContext(ctx, query, userID, conferenceID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}

		return nil, fmt.Errorf("failed to fetch booking by user and conference: %w", err)
	}

	return booking, nil
}

func (r *BookingRepository) UpdateRoom(ctx context.Context, bookingID, roomID int64) (*models.Booking, error) {
	query := `
		UPDATE bookings
		SET room_id = ?
		WHERE id = ? AND status != 'cancelled'
		RETURNING id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
	`

	updated, err := scanBooking(r.db.QueryRowContext(ctx, query, roomID, bookingID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to update booking room: %w", err)
	}

	return updated, nil
}

func (r *BookingRepository) Cancel(ctx context.Context, bookingID int64) (*models.Booking, error) {
	query := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status != 'cancelled'
		RETURNING id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
	`

	cancelled, err := scanBooking(r.db.QueryRowContext(ctx, query, bookingID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrBookingNotFound
		}
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	return cancelled, nil
}

func (r *BookingRepository) CancelAndCancelPendingOutgoingRequests(
	ctx context.Context,
	bookingID int64,
	userID int64,
) (*models.Booking, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to begin cancellation transaction: %w", err)
	}

	rolledBack := false
	rollback := func() {
		if rolledBack {
			return
		}
		_ = tx.Rollback()
		rolledBack = true
	}
	defer rollback()

	cancelQuery := `
		UPDATE bookings
		SET status = 'cancelled', cancelled_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND status != 'cancelled'
		RETURNING id, room_id, user_id, conference_id, status, privacy_setting, notes, created_at, confirmed_at, cancelled_at
	`

	cancelledBooking, err := scanBooking(tx.QueryRowContext(ctx, cancelQuery, bookingID, userID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, repository.ErrBookingNotFound
		}
		return nil, 0, fmt.Errorf("failed to cancel booking transactionally: %w", err)
	}

	cancelOutgoingQuery := `
		UPDATE roommate_requests
		SET status = 'cancelled'
		WHERE requester_id = ?
		  AND status = 'pending'
		  AND room_id IN (
		    SELECT id FROM rooms WHERE conference_id = ?
		  )
	`

	result, err := tx.ExecContext(ctx, cancelOutgoingQuery, userID, cancelledBooking.ConferenceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to cancel pending outgoing roommate requests transactionally: %w", err)
	}

	cancelledRequestCount, err := result.RowsAffected()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read cancelled roommate request count: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("failed to commit cancellation transaction: %w", err)
	}
	rolledBack = true

	return cancelledBooking, cancelledRequestCount, nil
}

func (r *BookingRepository) listBookings(ctx context.Context, query string, arg any) ([]*models.Booking, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookings: %w", err)
	}
	defer func() { _ = rows.Close() }()

	bookings := make([]*models.Booking, 0)
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking row: %w", err)
		}

		bookings = append(bookings, booking)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate booking rows: %w", err)
	}

	slog.Debug("bookings listed", "count", len(bookings))
	return bookings, nil
}

func scanBooking(s scanner) (*models.Booking, error) {
	var booking models.Booking
	var notes sql.NullString
	var confirmedAt sql.NullTime
	var cancelledAt sql.NullTime

	err := s.Scan(
		&booking.ID,
		&booking.RoomID,
		&booking.UserID,
		&booking.ConferenceID,
		&booking.Status,
		&booking.PrivacySetting,
		&notes,
		&booking.CreatedAt,
		&confirmedAt,
		&cancelledAt,
	)
	if err != nil {
		return nil, err
	}

	booking.Notes = notes.String
	if confirmedAt.Valid {
		booking.ConfirmedAt = &confirmedAt.Time
	}
	if cancelledAt.Valid {
		booking.CancelledAt = &cancelledAt.Time
	}

	return &booking, nil
}

func isBookingUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: bookings.user_id, bookings.conference_id")
}
