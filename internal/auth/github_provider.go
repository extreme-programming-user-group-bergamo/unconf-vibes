package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	githubDeviceCodeURL  = "https://github.com/login/device/code"
	githubAccessTokenURL = "https://github.com/login/oauth/access_token"
	githubUserURL        = "https://api.github.com/user"
	githubEmailsURL      = "https://api.github.com/user/emails"
)

type DeviceAuthorization struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type OAuthAccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GitHubProfile struct {
	GitHubID    string
	Email       string
	DisplayName string
}

type DeviceFlowError struct {
	Code        string
	Description string
	Interval    int
}

func (e *DeviceFlowError) Error() string {
	if e.Description == "" {
		return e.Code
	}

	return e.Code + ": " + e.Description
}

type GitHubProvider interface {
	StartDeviceFlow(ctx context.Context) (*DeviceAuthorization, error)
	ExchangeDeviceCode(ctx context.Context, deviceCode string) (*OAuthAccessToken, error)
	FetchProfile(ctx context.Context, accessToken string) (*GitHubProfile, error)
}

type HTTPGitHubProvider struct {
	client       *http.Client
	clientID     string
	clientSecret string
	scope        string
}

func NewHTTPGitHubProvider(clientID, clientSecret string, timeout time.Duration) *HTTPGitHubProvider {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &HTTPGitHubProvider{
		client:       &http.Client{Timeout: timeout},
		clientID:     clientID,
		clientSecret: clientSecret,
		scope:        "read:user user:email",
	}
}

func (p *HTTPGitHubProvider) StartDeviceFlow(ctx context.Context) (*DeviceAuthorization, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("scope", p.scope)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, githubDeviceCodeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create github device flow request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to call github device flow endpoint: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	payload := DeviceAuthorization{}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode github device flow response: %w", err)
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("github device flow returned status %d", response.StatusCode)
	}

	return &payload, nil
}

func (p *HTTPGitHubProvider) ExchangeDeviceCode(ctx context.Context, deviceCode string) (*OAuthAccessToken, error) {
	form := url.Values{}
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("device_code", deviceCode)
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, githubAccessTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create github token exchange request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to call github token exchange endpoint: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	var payload struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
		Interval         int    `json:"interval"`
	}

	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode github token exchange response: %w", err)
	}

	if payload.AccessToken != "" {
		return &OAuthAccessToken{
			AccessToken: payload.AccessToken,
			TokenType:   payload.TokenType,
			Scope:       payload.Scope,
		}, nil
	}

	if payload.Error != "" {
		return nil, &DeviceFlowError{Code: payload.Error, Description: payload.ErrorDescription, Interval: payload.Interval}
	}

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("github token exchange returned status %d", response.StatusCode)
	}

	return nil, fmt.Errorf("github token exchange returned no access token")
}

func (p *HTTPGitHubProvider) FetchProfile(ctx context.Context, accessToken string) (*GitHubProfile, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, githubUserURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create github user request: %w", err)
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to call github user endpoint: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("github user endpoint returned status %d", response.StatusCode)
	}

	var userPayload struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(response.Body).Decode(&userPayload); err != nil {
		return nil, fmt.Errorf("failed to decode github user response: %w", err)
	}

	email := userPayload.Email
	if email == "" {
		email, err = p.fetchPrimaryEmail(ctx, accessToken)
		if err != nil {
			return nil, err
		}
	}

	displayName := userPayload.Name
	if displayName == "" {
		displayName = userPayload.Login
	}

	return &GitHubProfile{
		GitHubID:    strconv.FormatInt(userPayload.ID, 10),
		Email:       email,
		DisplayName: displayName,
	}, nil
}

func (p *HTTPGitHubProvider) fetchPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, githubEmailsURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create github emails request: %w", err)
	}

	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+accessToken)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := p.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to call github emails endpoint: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("github emails endpoint returned status %d", response.StatusCode)
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.NewDecoder(response.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode github emails response: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified && email.Email != "" {
			return email.Email, nil
		}
	}

	for _, email := range emails {
		if email.Verified && email.Email != "" {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("github account has no verified email")
}

var _ GitHubProvider = (*HTTPGitHubProvider)(nil)
