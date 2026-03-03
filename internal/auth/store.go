package auth

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/zalando/go-keyring"
)

// ErrNotAuthenticated is returned when no valid auth token is available.
var ErrNotAuthenticated = errors.New("not authenticated: please run 'unconf login' to sign in")

// TokenStore defines the interface for secure token persistence.
type TokenStore interface {
	SaveTokens(accessToken, refreshToken string) error
	GetAccessToken() (string, error)
	GetRefreshToken() (string, error)
	ClearTokens() error
	HasValidToken() bool
}

const (
	keyAccessToken  = "access_token"
	keyRefreshToken = "refresh_token"
)

// KeyringTokenStore implements TokenStore using the OS keychain via go-keyring.
type KeyringTokenStore struct {
	serviceName string
}

// NewKeyringTokenStore creates a new KeyringTokenStore for the given service name.
func NewKeyringTokenStore(serviceName string) *KeyringTokenStore {
	return &KeyringTokenStore{serviceName: serviceName}
}

// SaveTokens stores the access and refresh tokens in the OS keychain.
func (s *KeyringTokenStore) SaveTokens(accessToken, refreshToken string) error {
	if err := keyring.Set(s.serviceName, keyAccessToken, accessToken); err != nil {
		return fmt.Errorf("failed to save access token to keyring: %w", err)
	}

	if err := keyring.Set(s.serviceName, keyRefreshToken, refreshToken); err != nil {
		return fmt.Errorf("failed to save refresh token to keyring: %w", err)
	}

	slog.Info("tokens saved to keyring", "service", s.serviceName)

	return nil
}

// GetAccessToken retrieves the access token from the OS keychain.
func (s *KeyringTokenStore) GetAccessToken() (string, error) {
	token, err := keyring.Get(s.serviceName, keyAccessToken)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", fmt.Errorf("failed to get access token: %w", ErrNotAuthenticated)
		}

		return "", fmt.Errorf("failed to get access token from keyring: %w", err)
	}

	return token, nil
}

// GetRefreshToken retrieves the refresh token from the OS keychain.
func (s *KeyringTokenStore) GetRefreshToken() (string, error) {
	token, err := keyring.Get(s.serviceName, keyRefreshToken)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", fmt.Errorf("failed to get refresh token: %w", ErrNotAuthenticated)
		}

		return "", fmt.Errorf("failed to get refresh token from keyring: %w", err)
	}

	return token, nil
}

// ClearTokens removes all stored tokens from the OS keychain.
func (s *KeyringTokenStore) ClearTokens() error {
	var errs []error

	if err := keyring.Delete(s.serviceName, keyAccessToken); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("failed to clear access token from keyring: %w", err))
	}

	if err := keyring.Delete(s.serviceName, keyRefreshToken); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		errs = append(errs, fmt.Errorf("failed to clear refresh token from keyring: %w", err))
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	slog.Info("tokens cleared from keyring", "service", s.serviceName)

	return nil
}

// HasValidToken checks whether an access token exists in the keychain.
func (s *KeyringTokenStore) HasValidToken() bool {
	token, err := keyring.Get(s.serviceName, keyAccessToken)
	return err == nil && token != ""
}

// RequireAuth checks if a valid token exists and returns ErrNotAuthenticated if not.
func RequireAuth(store TokenStore) error {
	if !store.HasValidToken() {
		return ErrNotAuthenticated
	}

	return nil
}
