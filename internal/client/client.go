package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

// Sentinel errors for client-side auth flow states.
var (
	ErrAuthorizationPending = errors.New("authorization pending")
	ErrSlowDown             = errors.New("slow down")
	ErrAccessDenied         = errors.New("access denied")
	ErrExpiredDeviceCode    = errors.New("device code expired")
	ErrRoomFull             = errors.New("room is at capacity")
	ErrAlreadyBooked        = errors.New("user already has a booking for this conference")
	ErrRoomNotFound         = errors.New("room not found")
	ErrTargetUserNotFound   = errors.New("target user not found")
	ErrTargetAlreadyBooked  = errors.New("target user already has a booking for this conference")
	ErrRequestPending       = errors.New("roommate request is already pending")
)

// DeviceFlowResponse represents the response from POST /auth/device.
type DeviceFlowResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse represents the response from POST /auth/token on success.
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user object within TokenResponse.
type UserResponse struct {
	ID             int64  `json:"id"`
	GitHubID       string `json:"github_id"`
	Email          string `json:"email"`
	DisplayName    string `json:"display_name"`
	PrivacySetting string `json:"privacy_setting"`
}

// UpdateProfileRequest represents the request body for PUT /users/me.
type UpdateProfileRequest struct {
	DisplayName    *string `json:"display_name,omitempty"`
	PrivacySetting *string `json:"privacy_setting,omitempty"`
}

// PendingResponse represents the 202 response from POST /auth/token while pending.
type PendingResponse struct {
	Status   string `json:"status"`
	Interval int    `json:"interval"`
}

// APIError represents an error response from the backend API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// apiErrorEnvelope mirrors the backend ErrorResponse shape.
type apiErrorEnvelope struct {
	Error APIError `json:"error"`
}

// Client is the HTTP client for CLI-to-API communication.
type Client struct {
	baseURL string
	http    *resty.Client
}

// NewClient creates a new API client targeting the given base URL.
func NewClient(baseURL string) *Client {
	r := resty.New().
		SetBaseURL(baseURL).
		SetHeader("Content-Type", "application/json")

	return &Client{
		baseURL: baseURL,
		http:    r,
	}
}

// StartDeviceFlow initiates the GitHub device flow via POST /auth/device.
func (c *Client) StartDeviceFlow(ctx context.Context) (*DeviceFlowResponse, error) {
	slog.Info("starting device flow", "endpoint", c.baseURL+"/auth/device")

	var result DeviceFlowResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&errEnvelope).
		Post("/auth/device")
	if err != nil {
		return nil, fmt.Errorf("failed to start device flow: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to start device flow: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}

// ExchangeDeviceCode polls POST /auth/token for the token exchange result.
// Returns ErrAuthorizationPending on 202 with authorization_pending status.
// Returns ErrSlowDown on 202 with slow_down status.
func (c *Client) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*TokenResponse, error) {
	body := map[string]string{"device_code": deviceCode}

	var tokenResp TokenResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&tokenResp).
		SetError(&errEnvelope).
		Post("/auth/token")
	if err != nil {
		return nil, fmt.Errorf("failed to exchange device code: %w", err)
	}

	if resp.StatusCode() == http.StatusAccepted {
		var pending PendingResponse
		if parseErr := json.Unmarshal(resp.Body(), &pending); parseErr == nil {
			switch pending.Status {
			case "slow_down":
				return nil, ErrSlowDown
			default:
				return nil, ErrAuthorizationPending
			}
		}

		return nil, ErrAuthorizationPending
	}

	if resp.StatusCode() == http.StatusBadRequest {
		return nil, fmt.Errorf("failed to exchange device code: %s: %w", errEnvelope.Error.Message, mapAPIErrorCode(errEnvelope.Error.Code))
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to exchange device code: %s: %w", errEnvelope.Error.Message, mapAPIErrorCode(errEnvelope.Error.Code))
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to exchange device code: unexpected HTTP %d", resp.StatusCode())
	}

	return &tokenResp, nil
}

// RefreshToken rotates an existing refresh token via POST /auth/refresh.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	body := map[string]string{"refresh_token": refreshToken}

	var tokenResp TokenResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetBody(body).
		SetResult(&tokenResp).
		SetError(&errEnvelope).
		Post("/auth/refresh")
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to refresh token: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &tokenResp, nil
}

// RevokeToken revokes the given access token via POST /auth/revoke.
func (c *Client) RevokeToken(ctx context.Context, accessToken string) error {
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetError(&errEnvelope).
		Post("/auth/revoke")
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("failed to revoke token: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return nil
}

// ErrUnauthorized is returned when the API responds with 401.
var ErrUnauthorized = errors.New("unauthorized")

// ErrConferenceNotFound is returned when the requested conference slug does not exist.
var ErrConferenceNotFound = errors.New("conference not found")

// ConferenceResponse represents a conference returned by the API.
type ConferenceResponse struct {
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Location      string `json:"location"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Capacity      int    `json:"capacity"`
	AttendeeCount int    `json:"attendee_count"`
	Status        string `json:"status"`
}

// ListConferences fetches all conferences via GET /conferences.
func (c *Client) ListConferences(ctx context.Context) ([]ConferenceResponse, error) {
	var result []ConferenceResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/conferences")
	if err != nil {
		return nil, fmt.Errorf("failed to list conferences: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to list conferences: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	if result == nil {
		result = []ConferenceResponse{}
	}

	return result, nil
}

// GetConference fetches a single conference by slug via GET /conferences/{slug}.
func (c *Client) GetConference(ctx context.Context, slug string) (*ConferenceResponse, error) {
	var result ConferenceResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetResult(&result).
		SetError(&errEnvelope).
		Get("/conferences/" + url.PathEscape(slug))
	if err != nil {
		return nil, fmt.Errorf("failed to get conference: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, ErrConferenceNotFound
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get conference: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &result, nil
}

// ErrSessionExpired is returned when the access token is expired and the refresh token
// is also invalid/expired/revoked, requiring a full re-login.
var ErrSessionExpired = errors.New("session expired: please run 'unconf login' to re-authenticate")

// GetMe fetches the current authenticated user's profile via GET /users/me.
func (c *Client) GetMe(ctx context.Context, accessToken string) (*UserResponse, error) {
	var user UserResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetResult(&user).
		SetError(&errEnvelope).
		Get("/users/me")
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to get user profile: %w", ErrUnauthorized)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get user profile: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &user, nil
}

// UpdateMe updates the authenticated user's profile via PUT /users/me.
func (c *Client) UpdateMe(ctx context.Context, accessToken string, input UpdateProfileRequest) (*UserResponse, error) {
	var user UserResponse
	var errEnvelope apiErrorEnvelope

	resp, err := c.http.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetBody(input).
		SetResult(&user).
		SetError(&errEnvelope).
		Put("/users/me")
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, fmt.Errorf("failed to update user profile: %w", ErrUnauthorized)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to update user profile: %s (HTTP %d)", errEnvelope.Error.Message, resp.StatusCode())
	}

	return &user, nil
}

// mapAPIErrorCode maps backend error codes to client-side sentinel errors.
func mapAPIErrorCode(code string) error {
	switch code {
	case "access_denied":
		return ErrAccessDenied
	case "expired_token":
		return ErrExpiredDeviceCode
	case "invalid_device_code":
		return fmt.Errorf("invalid device code")
	default:
		return fmt.Errorf("api error: %s", code)
	}
}
