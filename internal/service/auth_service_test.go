package service

import (
	"context"
	"testing"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testProvider struct {
	exchangeFn func(ctx context.Context, deviceCode string) (*auth.OAuthAccessToken, error)
	profileFn  func(ctx context.Context, accessToken string) (*auth.GitHubProfile, error)
}

func (p *testProvider) StartDeviceFlow(ctx context.Context) (*auth.DeviceAuthorization, error) {
	return &auth.DeviceAuthorization{}, nil
}

func (p *testProvider) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*auth.OAuthAccessToken, error) {
	if p.exchangeFn == nil {
		return nil, nil
	}

	return p.exchangeFn(ctx, deviceCode)
}

func (p *testProvider) FetchProfile(ctx context.Context, accessToken string) (*auth.GitHubProfile, error) {
	if p.profileFn == nil {
		return nil, nil
	}

	return p.profileFn(ctx, accessToken)
}

type testTokenIssuer struct{}

func (i *testTokenIssuer) IssueAccessToken(ctx context.Context, input auth.AccessTokenInput) (string, error) {
	return "issued-access-token", nil
}

func (i *testTokenIssuer) GenerateTokenID() (string, error) {
	return "jti-123", nil
}

func (i *testTokenIssuer) GenerateRefreshToken() (string, string, error) {
	return "refresh-token", "refresh-hash", nil
}

type testRefreshRepository struct {
	createFn func(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error)
	getFn    func(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
}

func (r *testRefreshRepository) Create(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error) {
	if r.createFn == nil {
		return nil, nil
	}

	return r.createFn(ctx, session)
}

func (r *testRefreshRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	if r.getFn == nil {
		return nil, nil
	}

	return r.getFn(ctx, tokenHash)
}

func (r *testRefreshRepository) Rotate(ctx context.Context, currentSessionID int64, replacement *models.RefreshSession) (*models.RefreshSession, error) {
	return &models.RefreshSession{ID: 100, UserID: replacement.UserID}, nil
}

func (r *testRefreshRepository) RevokeByID(_ context.Context, _ int64) error {
	return nil
}

func TestAuthServiceExchangeDeviceCodeMapsPendingError(t *testing.T) {
	svc, err := NewAuthService(
		&testProvider{exchangeFn: func(ctx context.Context, deviceCode string) (*auth.OAuthAccessToken, error) {
			return nil, &auth.DeviceFlowError{Code: "authorization_pending", Interval: 8}
		}},
		&testTokenIssuer{},
		&repository.MockUserRepository{},
		&testRefreshRepository{},
		time.Hour,
		24*time.Hour,
	)
	require.NoError(t, err)

	_, err = svc.ExchangeDeviceCode(context.Background(), "device-code")
	require.Error(t, err)

	var pendingErr *PendingAuthError
	assert.ErrorAs(t, err, &pendingErr)
	assert.Equal(t, 8, pendingErr.Interval)
	assert.ErrorIs(t, err, ErrAuthorizationPending)
}

func TestAuthServiceExchangeDeviceCodeSuccessCreatesUserAndSession(t *testing.T) {
	createdUser := &models.User{ID: 42, GitHubID: "12345", Email: "new@example.com", DisplayName: "octocat"}
	refreshSession := &models.RefreshSession{ID: 99, UserID: createdUser.ID}

	userRepo := &repository.MockUserRepository{
		GetByGitHubFunc: func(ctx context.Context, githubID string) (*models.User, error) {
			return nil, repository.ErrUserNotFound
		},
		CreateFunc: func(ctx context.Context, user *models.User) (*models.User, error) {
			return createdUser, nil
		},
	}

	refreshRepo := &testRefreshRepository{
		createFn: func(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error) {
			assert.Equal(t, createdUser.ID, session.UserID)
			assert.Equal(t, "refresh-hash", session.TokenHash)
			return refreshSession, nil
		},
	}

	provider := &testProvider{
		exchangeFn: func(ctx context.Context, deviceCode string) (*auth.OAuthAccessToken, error) {
			return &auth.OAuthAccessToken{AccessToken: "gh-token"}, nil
		},
		profileFn: func(ctx context.Context, accessToken string) (*auth.GitHubProfile, error) {
			return &auth.GitHubProfile{GitHubID: "12345", Email: "new@example.com", DisplayName: "octocat"}, nil
		},
	}

	svc, err := NewAuthService(provider, &testTokenIssuer{}, userRepo, refreshRepo, time.Hour, 24*time.Hour)
	require.NoError(t, err)

	result, err := svc.ExchangeDeviceCode(context.Background(), "device-code")
	require.NoError(t, err)
	assert.Equal(t, "issued-access-token", result.AccessToken)
	assert.Equal(t, "refresh-token", result.RefreshToken)
	assert.Equal(t, createdUser.ID, result.User.ID)
}
