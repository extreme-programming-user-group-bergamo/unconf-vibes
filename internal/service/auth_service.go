package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/katurdays/unconf/internal/auth"
	"github.com/katurdays/unconf/internal/models"
	"github.com/katurdays/unconf/internal/repository"
)

const (
	defaultTokenType = "Bearer"
	defaultIssuer    = "unconf-api"
	defaultAudience  = "unconf-cli"
)

type DeviceFlowProvider interface {
	StartDeviceFlow(ctx context.Context) (*auth.DeviceAuthorization, error)
	ExchangeDeviceCode(ctx context.Context, deviceCode string) (*auth.OAuthAccessToken, error)
	FetchProfile(ctx context.Context, accessToken string) (*auth.GitHubProfile, error)
}

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, input auth.AccessTokenInput) (string, error)
	GenerateTokenID() (string, error)
	GenerateRefreshToken() (raw string, hash string, err error)
}

type AuthResult struct {
	AccessToken  string            `json:"access_token"`
	TokenType    string            `json:"token_type"`
	ExpiresIn    int               `json:"expires_in"`
	RefreshToken string            `json:"refresh_token"`
	User         AuthenticatedUser `json:"user"`
}

type AuthenticatedUser struct {
	ID          int64  `json:"id"`
	GitHubID    string `json:"github_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type AuthService struct {
	provider     DeviceFlowProvider
	tokenIssuer  TokenIssuer
	userRepo     repository.UserRepository
	refreshRepo  repository.RefreshSessionRepository
	accessTTL    time.Duration
	refreshTTL   time.Duration
	issuer       string
	audience     string
	nowFunc      func() time.Time
	tokenTypeOut string
}

func NewAuthService(
	provider DeviceFlowProvider,
	tokenIssuer TokenIssuer,
	userRepo repository.UserRepository,
	refreshRepo repository.RefreshSessionRepository,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) (*AuthService, error) {
	if provider == nil {
		return nil, fmt.Errorf("failed to initialize auth service: provider is nil")
	}

	if tokenIssuer == nil {
		return nil, fmt.Errorf("failed to initialize auth service: token issuer is nil")
	}

	if userRepo == nil {
		return nil, fmt.Errorf("failed to initialize auth service: user repository is nil")
	}

	if refreshRepo == nil {
		return nil, fmt.Errorf("failed to initialize auth service: refresh session repository is nil")
	}

	if accessTTL <= 0 {
		return nil, fmt.Errorf("failed to initialize auth service: access ttl must be positive")
	}

	if refreshTTL <= 0 {
		return nil, fmt.Errorf("failed to initialize auth service: refresh ttl must be positive")
	}

	return &AuthService{
		provider:     provider,
		tokenIssuer:  tokenIssuer,
		userRepo:     userRepo,
		refreshRepo:  refreshRepo,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
		issuer:       defaultIssuer,
		audience:     defaultAudience,
		nowFunc:      time.Now,
		tokenTypeOut: defaultTokenType,
	}, nil
}

func (s *AuthService) StartDeviceFlow(ctx context.Context) (*auth.DeviceAuthorization, error) {
	deviceAuth, err := s.provider.StartDeviceFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start device flow: %w", err)
	}

	slog.Info("github device flow started", "verification_uri", deviceAuth.VerificationURI, "interval", deviceAuth.Interval)
	return deviceAuth, nil
}

func (s *AuthService) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*AuthResult, error) {
	if strings.TrimSpace(deviceCode) == "" {
		return nil, fmt.Errorf("failed to exchange device code: %w", ErrInvalidDeviceCode)
	}

	token, err := s.provider.ExchangeDeviceCode(ctx, deviceCode)
	if err != nil {
		mapped := mapDeviceFlowError(err)
		if mapped != nil {
			return nil, mapped
		}

		return nil, fmt.Errorf("failed to exchange device code: %w", err)
	}

	profile, err := s.provider.FetchProfile(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github profile: %w", err)
	}

	user, err := s.upsertUser(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to persist authenticated user: %w", err)
	}

	result, err := s.issueSessionTokens(ctx, user, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to issue session tokens: %w", err)
	}

	slog.Info("device code exchange succeeded", "user_id", user.ID, "github_id", user.GitHubID)
	return result, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrInvalidRefreshToken)
	}

	session, err := s.refreshRepo.GetByTokenHash(ctx, auth.HashRefreshToken(refreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrRefreshSessionNotFound) {
			return nil, fmt.Errorf("failed to refresh token: %w", ErrInvalidRefreshToken)
		}

		return nil, fmt.Errorf("failed to lookup refresh session: %w", err)
	}

	now := s.nowFunc().UTC()
	if session.RevokedAt != nil || session.ReplacedByID != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrRevokedRefreshToken)
	}

	if !session.ExpiresAt.After(now) {
		return nil, fmt.Errorf("failed to refresh token: %w", ErrExpiredRefreshToken)
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user for refresh: %w", err)
	}

	result, err := s.issueSessionTokens(ctx, user, session)
	if err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	slog.Info("refresh token rotated", "user_id", user.ID, "session_id", session.ID)
	return result, nil
}

func (s *AuthService) issueSessionTokens(ctx context.Context, user *models.User, currentSession *models.RefreshSession) (*AuthResult, error) {
	now := s.nowFunc().UTC()

	jti, err := s.tokenIssuer.GenerateTokenID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token id: %w", err)
	}

	rawRefreshToken, hashedRefreshToken, err := s.tokenIssuer.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	replacement := &models.RefreshSession{
		UserID:        user.ID,
		TokenHash:     hashedRefreshToken,
		ExpiresAt:     now.Add(s.refreshTTL),
		IssuedAt:      now,
		LastAccessJTI: jti,
	}

	var persisted *models.RefreshSession
	if currentSession == nil {
		persisted, err = s.refreshRepo.Create(ctx, replacement)
	} else {
		persisted, err = s.refreshRepo.Rotate(ctx, currentSession.ID, replacement)
		if errors.Is(err, repository.ErrRefreshSessionNotFound) {
			return nil, fmt.Errorf("failed to rotate refresh token: %w", ErrInvalidRefreshToken)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to persist refresh session: %w", err)
	}

	accessToken, err := s.tokenIssuer.IssueAccessToken(ctx, auth.AccessTokenInput{
		UserID:     user.ID,
		SessionID:  persisted.ID,
		Issuer:     s.issuer,
		Audience:   s.audience,
		NotBefore:  now,
		IssuedAt:   now,
		JTI:        jti,
		ExpiryTime: now.Add(s.accessTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		TokenType:    s.tokenTypeOut,
		ExpiresIn:    int(s.accessTTL.Seconds()),
		RefreshToken: rawRefreshToken,
		User: AuthenticatedUser{
			ID:          user.ID,
			GitHubID:    user.GitHubID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}, nil
}

// RevokeSession revokes a refresh session by its ID.
func (s *AuthService) RevokeSession(ctx context.Context, sessionID int64) error {
	if err := s.refreshRepo.RevokeByID(ctx, sessionID); err != nil {
		if errors.Is(err, repository.ErrRefreshSessionNotFound) {
			return fmt.Errorf("failed to revoke session: %w", ErrSessionNotFound)
		}

		return fmt.Errorf("failed to revoke session: %w", err)
	}

	slog.Info("session revoked", "session_id", sessionID)

	return nil
}

func (s *AuthService) upsertUser(ctx context.Context, profile *auth.GitHubProfile) (*models.User, error) {
	if profile == nil {
		return nil, fmt.Errorf("failed to upsert user: github profile is nil")
	}

	user, err := s.userRepo.GetByGitHubID(ctx, profile.GitHubID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			created, createErr := s.userRepo.Create(ctx, &models.User{
				GitHubID:       profile.GitHubID,
				Email:          profile.Email,
				DisplayName:    profile.DisplayName,
				PrivacySetting: "public",
			})
			if createErr != nil {
				return nil, fmt.Errorf("failed to create user: %w", createErr)
			}

			return created, nil
		}

		return nil, fmt.Errorf("failed to lookup user by github id: %w", err)
	}

	if user.Email == profile.Email && user.DisplayName == profile.DisplayName {
		return user, nil
	}

	user.Email = profile.Email
	user.DisplayName = profile.DisplayName

	updated, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updated, nil
}

func mapDeviceFlowError(err error) error {
	var flowErr *auth.DeviceFlowError
	if !errors.As(err, &flowErr) {
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(flowErr.Code)) {
	case "authorization_pending":
		return &PendingAuthError{Interval: flowErr.Interval, Cause: ErrAuthorizationPending}
	case "slow_down":
		return &PendingAuthError{Interval: flowErr.Interval, Cause: ErrSlowDown}
	case "access_denied":
		return fmt.Errorf("github denied authorization: %w", ErrAccessDenied)
	case "expired_token":
		return fmt.Errorf("github device token expired: %w", ErrExpiredDeviceCode)
	case "invalid_request", "incorrect_device_code", "bad_verification_code", "invalid_device_code":
		return fmt.Errorf("github device code is invalid: %w", ErrInvalidDeviceCode)
	default:
		return fmt.Errorf("github device flow error %q: %w", flowErr.Code, err)
	}
}

var _ TokenIssuer = (*auth.TokenService)(nil)
