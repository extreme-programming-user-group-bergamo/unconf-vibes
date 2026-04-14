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

var _ repository.RoomRepository = (*RoomRepository)(nil)

type RoomRepository struct {
	db *sql.DB
}

func NewRoomRepository(db *sql.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *models.Room) (*models.Room, error) {
	if room == nil {
		return nil, fmt.Errorf("failed to create room: room is nil")
	}

	query := `
		INSERT INTO rooms (conference_id, room_number, room_type, price_per_night, capacity)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, conference_id, room_number, room_type, price_per_night, capacity, created_at
	`

	created, err := scanRoom(r.db.QueryRowContext(ctx, query,
		room.ConferenceID, room.RoomNumber, room.RoomType,
		room.PricePerNight, room.Capacity,
	))
	if err != nil {
		if isRoomUniqueConstraintError(err) {
			return nil, repository.ErrRoomExists
		}

		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	slog.Debug("room created", "room_id", created.ID, "room_number", created.RoomNumber)
	return created, nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id int64) (*models.Room, error) {
	query := `
		SELECT id, conference_id, room_number, room_type, price_per_night, capacity, created_at
		FROM rooms
		WHERE id = ?
	`

	room, err := scanRoom(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRoomNotFound
		}

		return nil, fmt.Errorf("failed to fetch room by id: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) GetByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) (*models.Room, error) {
	query := `
		SELECT id, conference_id, room_number, room_type, price_per_night, capacity, created_at
		FROM rooms
		WHERE conference_id = ? AND room_number = ?
	`

	room, err := scanRoom(r.db.QueryRowContext(ctx, query, conferenceID, roomNumber))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRoomNotFound
		}

		return nil, fmt.Errorf("failed to fetch room by conference and number: %w", err)
	}

	return room, nil
}

func (r *RoomRepository) ListByConference(ctx context.Context, conferenceID int64) ([]*models.Room, error) {
	query := `
		SELECT id, conference_id, room_number, room_type, price_per_night, capacity, created_at
		FROM rooms
		WHERE conference_id = ?
		ORDER BY room_number
	`

	rows, err := r.db.QueryContext(ctx, query, conferenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms by conference: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rooms := make([]*models.Room, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room row: %w", err)
		}

		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate room rows: %w", err)
	}

	slog.Debug("rooms listed", "conference_id", conferenceID, "count", len(rooms))
	return rooms, nil
}

func (r *RoomRepository) UpdateByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string, room *models.Room) (*models.Room, error) {
	if room == nil {
		return nil, fmt.Errorf("failed to update room: room is nil")
	}

	query := `
		UPDATE rooms
		SET room_number = ?, room_type = ?, price_per_night = ?, capacity = ?
		WHERE conference_id = ? AND room_number = ?
		RETURNING id, conference_id, room_number, room_type, price_per_night, capacity, created_at
	`

	updated, err := scanRoom(r.db.QueryRowContext(
		ctx,
		query,
		room.RoomNumber,
		room.RoomType,
		room.PricePerNight,
		room.Capacity,
		conferenceID,
		roomNumber,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRoomNotFound
		}
		if isRoomUniqueConstraintError(err) {
			return nil, repository.ErrRoomExists
		}

		return nil, fmt.Errorf("failed to update room: %w", err)
	}

	slog.Debug("room updated", "room_id", updated.ID, "room_number", updated.RoomNumber)
	return updated, nil
}

func (r *RoomRepository) DeleteByConferenceAndNumber(ctx context.Context, conferenceID int64, roomNumber string) error {
	query := `
		DELETE FROM rooms
		WHERE conference_id = ? AND room_number = ?
	`

	result, err := r.db.ExecContext(ctx, query, conferenceID, roomNumber)
	if err != nil {
		if isRoomForeignKeyConstraintError(err) {
			return repository.ErrRoomHasBookings
		}

		return fmt.Errorf("failed to delete room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to delete room: %w", err)
	}
	if rowsAffected == 0 {
		return repository.ErrRoomNotFound
	}

	slog.Debug("room deleted", "conference_id", conferenceID, "room_number", roomNumber)
	return nil
}

func scanRoom(s scanner) (*models.Room, error) {
	var room models.Room

	err := s.Scan(
		&room.ID,
		&room.ConferenceID,
		&room.RoomNumber,
		&room.RoomType,
		&room.PricePerNight,
		&room.Capacity,
		&room.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func isRoomUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: rooms.conference_id, rooms.room_number")
}

func isRoomForeignKeyConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "foreign key constraint failed")
}
