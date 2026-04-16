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
	createFn       func(ctx context.Context, session *models.RefreshSession) (*models.RefreshSession, error)
	getFn          func(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
	listFn         func(ctx context.Context, userID int64) ([]*models.RefreshSession, error)
	revokeForUser  func(ctx context.Context, userID int64, sessionID int64) error
	revokeOthersFn func(ctx context.Context, userID int64, keepSessionID int64) (int64, error)
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

func (r *testRefreshRepository) ListActiveByUser(ctx context.Context, userID int64) ([]*models.RefreshSession, error) {
	if r.listFn == nil {
		return nil, nil
	}

	return r.listFn(ctx, userID)
}

func (r *testRefreshRepository) RevokeByID(_ context.Context, _ int64) error {
	return nil
}

func (r *testRefreshRepository) RevokeByUserAndID(ctx context.Context, userID int64, sessionID int64) error {
	if r.revokeForUser != nil {
		return r.revokeForUser(ctx, userID, sessionID)
	}
	return nil
}

func (r *testRefreshRepository) RevokeAllByUserExceptSession(ctx context.Context, userID int64, keepSessionID int64) (int64, error) {
	if r.revokeOthersFn != nil {
		return r.revokeOthersFn(ctx, userID, keepSessionID)
	}
	return 0, nil
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

func TestAuthServiceListActiveSessions(t *testing.T) {
	now := time.Now().UTC()

	svc, err := NewAuthService(
		&testProvider{},
		&testTokenIssuer{},
		&repository.MockUserRepository{},
		&testRefreshRepository{
			listFn: func(_ context.Context, userID int64) ([]*models.RefreshSession, error) {
				assert.Equal(t, int64(7), userID)
				return []*models.RefreshSession{
					{
						ID:         11,
						ClientInfo: "ua=other",
						CreatedAt:  now.Add(-time.Hour),
						IssuedAt:   now.Add(-time.Hour),
					},
					{
						ID:         22,
						ClientInfo: "ua=current",
						CreatedAt:  now,
						IssuedAt:   now,
					},
				}, nil
			},
		},
		time.Hour,
		24*time.Hour,
	)
	require.NoError(t, err)

	sessions, err := svc.ListActiveSessions(context.Background(), 7, 22)
	require.NoError(t, err)
	require.Len(t, sessions, 2)
	assert.Equal(t, int64(22), sessions[0].ID)
	assert.True(t, sessions[0].Current)
	assert.Equal(t, "ua=current", sessions[0].ClientMetadata)
}

func TestAuthServiceRevokeSessionForUserCurrentBlocked(t *testing.T) {
	svc, err := NewAuthService(
		&testProvider{},
		&testTokenIssuer{},
		&repository.MockUserRepository{},
		&testRefreshRepository{},
		time.Hour,
		24*time.Hour,
	)
	require.NoError(t, err)

	err = svc.RevokeSessionForUser(context.Background(), 7, 44, 44)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeCurrentSession)
}

func TestAuthServiceRevokeOtherSessions(t *testing.T) {
	svc, err := NewAuthService(
		&testProvider{},
		&testTokenIssuer{},
		&repository.MockUserRepository{},
		&testRefreshRepository{
			revokeOthersFn: func(_ context.Context, userID int64, keepSessionID int64) (int64, error) {
				assert.Equal(t, int64(7), userID)
				assert.Equal(t, int64(44), keepSessionID)
				return 2, nil
			},
		},
		time.Hour,
		24*time.Hour,
	)
	require.NoError(t, err)

	revokedCount, err := svc.RevokeOtherSessions(context.Background(), 7, 44)
	require.NoError(t, err)
	assert.Equal(t, int64(2), revokedCount)
}
