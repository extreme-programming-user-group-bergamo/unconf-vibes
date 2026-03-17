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

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("failed to create user: user is nil")
	}

	query := `
		INSERT INTO users (github_id, email, display_name, privacy_setting)
		VALUES (?, ?, ?, ?)
		RETURNING id, github_id, email, display_name, privacy_setting, created_at, updated_at
	`

	privacySetting := user.PrivacySetting
	if privacySetting == "" {
		privacySetting = "public"
	}

	created := models.User{}
	err := r.db.QueryRowContext(ctx, query, user.GitHubID, user.Email, user.DisplayName, privacySetting).Scan(
		&created.ID,
		&created.GitHubID,
		&created.Email,
		&created.DisplayName,
		&created.PrivacySetting,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, repository.ErrUserExists
		}

		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	slog.Debug("user created", "user_id", created.ID, "github_id", created.GitHubID)
	return &created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, github_id, email, display_name, privacy_setting, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user, err := scanUserRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	query := `
		SELECT id, github_id, email, display_name, privacy_setting, created_at, updated_at
		FROM users
		WHERE github_id = ?
	`

	user, err := scanUserRow(r.db.QueryRowContext(ctx, query, githubID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to fetch user by github_id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("failed to update user: user is nil")
	}

	query := `
		UPDATE users
		SET email = ?, display_name = ?, privacy_setting = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.DisplayName, user.PrivacySetting, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to read updated rows count: %w", err)
	}

	if affectedRows == 0 {
		return nil, repository.ErrUserNotFound
	}

	updated, err := r.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated user: %w", err)
	}

	slog.Debug("user updated", "user_id", user.ID, "github_id", updated.GitHubID)
	return updated, nil
}

func scanUserRow(row *sql.Row) (*models.User, error) {
	var user models.User
	var email sql.NullString
	var displayName sql.NullString
	var privacySetting sql.NullString

	err := row.Scan(
		&user.ID,
		&user.GitHubID,
		&email,
		&displayName,
		&privacySetting,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.Email = email.String
	user.DisplayName = displayName.String
	user.PrivacySetting = privacySetting.String

	return &user, nil
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed: users.github_id")
}

var _ repository.UserRepository = (*UserRepository)(nil)
