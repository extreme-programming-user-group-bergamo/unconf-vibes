package repository

import (
	"context"

	"github.com/katurdays/unconf/internal/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByGitHubID(ctx context.Context, githubID string) (*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
}

type RefreshSessionRepository interface {
	Create(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
	Rotate(ctx context.Context, currentSessionID int64, replacement *models.RefreshSession) (*models.RefreshSession, error)
	RevokeByID(ctx context.Context, sessionID int64) error
}

type ConferenceRepository interface {
	Create(ctx context.Context, conf *models.Conference) (*models.Conference, error)
	GetBySlug(ctx context.Context, slug string) (*models.Conference, error)
	List(ctx context.Context) ([]*models.Conference, error)
}
