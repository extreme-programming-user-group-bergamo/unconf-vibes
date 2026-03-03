package repository

import (
	"context"

	"github.com/katurdays/unconf/internal/models"
)

type MockUserRepository struct {
	CreateFunc      func(ctx context.Context, user *models.User) (*models.User, error)
	GetByIDFunc     func(ctx context.Context, id int64) (*models.User, error)
	GetByGitHubFunc func(ctx context.Context, githubID string) (*models.User, error)
	UpdateFunc      func(ctx context.Context, user *models.User) (*models.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if m.CreateFunc == nil {
		return nil, nil
	}

	return m.CreateFunc(ctx, user)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if m.GetByIDFunc == nil {
		return nil, nil
	}

	return m.GetByIDFunc(ctx, id)
}

func (m *MockUserRepository) GetByGitHubID(ctx context.Context, githubID string) (*models.User, error) {
	if m.GetByGitHubFunc == nil {
		return nil, nil
	}

	return m.GetByGitHubFunc(ctx, githubID)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	if m.UpdateFunc == nil {
		return nil, nil
	}

	return m.UpdateFunc(ctx, user)
}

var _ UserRepository = (*MockUserRepository)(nil)
