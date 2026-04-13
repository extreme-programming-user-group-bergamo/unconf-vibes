package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/katurdays/unconf/internal/auth"
)

// AuthenticatedClient wraps Client with automatic token management.
// It attaches the stored access token to requests and transparently
// refreshes expired tokens using the stored refresh token.
type AuthenticatedClient struct {
	client *Client
	store  auth.TokenStore
}

// NewAuthenticatedClient creates a new AuthenticatedClient.
func NewAuthenticatedClient(client *Client, store auth.TokenStore) *AuthenticatedClient {
	return &AuthenticatedClient{
		client: client,
		store:  store,
	}
}

// GetMe fetches the authenticated user's profile, automatically refreshing
// the access token on 401.
func (ac *AuthenticatedClient) GetMe(ctx context.Context) (*UserResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "GetMe")

	user, err := ac.client.GetMe(ctx, accessToken)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	// Access token rejected — attempt refresh
	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "GetMe")

	retryUser, retryErr := ac.client.GetMe(ctx, newAccessToken)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to get user profile after token refresh: %w", retryErr)
	}

	return retryUser, nil
}

// UpdateMe updates the authenticated user's profile, automatically refreshing
// the access token on 401.
func (ac *AuthenticatedClient) UpdateMe(ctx context.Context, input UpdateProfileRequest) (*UserResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "UpdateMe")

	user, err := ac.client.UpdateMe(ctx, accessToken, input)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	// Access token rejected — attempt refresh
	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "UpdateMe")

	retryUser, retryErr := ac.client.UpdateMe(ctx, newAccessToken, input)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to update user profile after token refresh: %w", retryErr)
	}

	return retryUser, nil
}

// CreateBooking creates a booking, automatically refreshing the access token on 401.
func (ac *AuthenticatedClient) CreateBooking(ctx context.Context, input CreateBookingRequest) (*BookingResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "CreateBooking")

	booking, err := ac.client.CreateBooking(ctx, accessToken, input)
	if err == nil {
		return booking, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "CreateBooking")

	retryBooking, retryErr := ac.client.CreateBooking(ctx, newAccessToken, input)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to create booking after token refresh: %w", retryErr)
	}

	return retryBooking, nil
}

// CreateRoommateRequest creates a roommate request, automatically refreshing the access token on 401.
func (ac *AuthenticatedClient) CreateRoommateRequest(ctx context.Context, input CreateRoommateRequestRequest) (*RoommateRequestResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "CreateRoommateRequest")

	request, err := ac.client.CreateRoommateRequest(ctx, accessToken, input)
	if err == nil {
		return request, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "CreateRoommateRequest")

	retryRequest, retryErr := ac.client.CreateRoommateRequest(ctx, newAccessToken, input)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to create roommate request after token refresh: %w", retryErr)
	}

	return retryRequest, nil
}

// ListBookings fetches authenticated user bookings, automatically refreshing the access token on 401.
func (ac *AuthenticatedClient) ListBookings(ctx context.Context) ([]BookingResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "ListBookings")

	bookings, err := ac.client.ListBookings(ctx, accessToken)
	if err == nil {
		return bookings, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "ListBookings")

	retryBookings, retryErr := ac.client.ListBookings(ctx, newAccessToken)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to list bookings after token refresh: %w", retryErr)
	}

	return retryBookings, nil
}

// ListRoommateRequests fetches roommate requests, automatically refreshing the access token on 401.
func (ac *AuthenticatedClient) ListRoommateRequests(ctx context.Context) ([]RoommateRequestResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "ListRoommateRequests")

	requests, err := ac.client.ListRoommateRequests(ctx, accessToken)
	if err == nil {
		return requests, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "ListRoommateRequests")

	retryRequests, retryErr := ac.client.ListRoommateRequests(ctx, newAccessToken)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to list roommate requests after token refresh: %w", retryErr)
	}

	return retryRequests, nil
}

// ListAttendees fetches conference attendees, automatically refreshing the access token on 401.
func (ac *AuthenticatedClient) ListAttendees(ctx context.Context, slug string) (*AttendeeListResponse, error) {
	accessToken, err := ac.store.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	slog.Debug("auth client: attempting authenticated request", "method", "ListAttendees")

	attendees, err := ac.client.ListAttendees(ctx, accessToken, slug)
	if err == nil {
		return attendees, nil
	}

	if !errors.Is(err, ErrUnauthorized) {
		return nil, err
	}

	newAccessToken, refreshErr := ac.tryRefresh(ctx)
	if refreshErr != nil {
		return nil, refreshErr
	}

	slog.Debug("auth client: retrying request after token refresh", "method", "ListAttendees")

	retryAttendees, retryErr := ac.client.ListAttendees(ctx, newAccessToken, slug)
	if retryErr != nil {
		return nil, fmt.Errorf("failed to list attendees after token refresh: %w", retryErr)
	}

	return retryAttendees, nil
}

// tryRefresh attempts to refresh the access token using the stored refresh token.
// On success, it saves new tokens and returns the new access token.
// On failure, it clears all tokens and returns ErrSessionExpired.
func (ac *AuthenticatedClient) tryRefresh(ctx context.Context) (string, error) {
	refreshToken, err := ac.store.GetRefreshToken()
	if err != nil {
		slog.Warn("token refresh failed, no refresh token available", "error", err)

		if clearErr := ac.store.ClearTokens(); clearErr != nil {
			slog.Error("failed to clear tokens after missing refresh token", "error", clearErr)
		}

		return "", fmt.Errorf("failed to refresh session: %w", ErrSessionExpired)
	}

	tokenResp, err := ac.client.RefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Warn("token refresh failed, clearing credentials", "error", err)

		if clearErr := ac.store.ClearTokens(); clearErr != nil {
			slog.Error("failed to clear tokens after refresh failure", "error", clearErr)
		}

		return "", fmt.Errorf("failed to refresh session: %w", ErrSessionExpired)
	}

	if err := ac.store.SaveTokens(tokenResp.AccessToken, tokenResp.RefreshToken); err != nil {
		return "", fmt.Errorf("failed to save refreshed tokens: %w", err)
	}

	slog.Info("access token refreshed successfully")

	return tokenResp.AccessToken, nil
}
