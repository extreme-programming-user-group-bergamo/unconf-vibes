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
