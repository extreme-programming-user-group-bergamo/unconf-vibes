package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

type RefreshSessionRepository struct {
	db *sql.DB
}

func NewRefreshSessionRepository(db *sql.DB) *RefreshSessionRepository {
	return &RefreshSessionRepository{db: db}
}

func (r *RefreshSessionRepository) Create(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error) {
	if session == nil {
		return nil, fmt.Errorf("failed to create refresh session: session is nil")
	}

	query := `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at, issued_at, replaced_by_id, revoked_at, last_access_jti)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, user_id, token_hash, expires_at, issued_at, replaced_by_id, revoked_at, last_access_jti, created_at, updated_at
	`

	created := models.RefreshSession{}
	err := r.db.QueryRowContext(
		ctx,
		query,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt.UTC(),
		session.IssuedAt.UTC(),
		session.ReplacedByID,
		session.RevokedAt,
		session.LastAccessJTI,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.IssuedAt,
		&created.ReplacedByID,
		&created.RevokedAt,
		&created.LastAccessJTI,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh session: %w", err)
	}

	return &created, nil
}

func (r *RefreshSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, issued_at, replaced_by_id, revoked_at, last_access_jti, created_at, updated_at
		FROM refresh_sessions
		WHERE token_hash = ?
	`

	session, err := scanRefreshSessionRow(r.db.QueryRowContext(ctx, query, tokenHash))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrRefreshSessionNotFound
		}

		return nil, fmt.Errorf("failed to fetch refresh session by token hash: %w", err)
	}

	return session, nil
}

func (r *RefreshSessionRepository) Rotate(ctx context.Context, currentSessionID int64, replacement *models.RefreshSession) (*models.RefreshSession, error) {
	if replacement == nil {
		return nil, fmt.Errorf("failed to rotate refresh session: replacement is nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin refresh session rotation transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	created, err := insertRefreshSessionTx(ctx, tx, replacement)
	if err != nil {
		return nil, err
	}

	updateQuery := `
		UPDATE refresh_sessions
		SET revoked_at = ?, replaced_by_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND revoked_at IS NULL
	`

	result, err := tx.ExecContext(ctx, updateQuery, time.Now().UTC(), created.ID, currentSessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke current refresh session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check revoked refresh session rows: %w", err)
	}

	if affected == 0 {
		return nil, repository.ErrRefreshSessionNotFound
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, fmt.Errorf("failed to commit refresh session rotation transaction: %w", commitErr)
	}

	return created, nil
}

func insertRefreshSessionTx(ctx context.Context, tx *sql.Tx, session *models.RefreshSession) (*models.RefreshSession, error) {
	query := `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at, issued_at, replaced_by_id, revoked_at, last_access_jti)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, user_id, token_hash, expires_at, issued_at, replaced_by_id, revoked_at, last_access_jti, created_at, updated_at
	`

	created := models.RefreshSession{}
	err := tx.QueryRowContext(
		ctx,
		query,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt.UTC(),
		session.IssuedAt.UTC(),
		session.ReplacedByID,
		session.RevokedAt,
		session.LastAccessJTI,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.IssuedAt,
		&created.ReplacedByID,
		&created.RevokedAt,
		&created.LastAccessJTI,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replacement refresh session: %w", err)
	}

	return &created, nil
}

func scanRefreshSessionRow(row *sql.Row) (*models.RefreshSession, error) {
	var session models.RefreshSession
	var replacedByID sql.NullInt64
	var revokedAt sql.NullTime

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.ExpiresAt,
		&session.IssuedAt,
		&replacedByID,
		&revokedAt,
		&session.LastAccessJTI,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if replacedByID.Valid {
		session.ReplacedByID = &replacedByID.Int64
	}

	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return &session, nil
}

var _ repository.RefreshSessionRepository = (*RefreshSessionRepository)(nil)
