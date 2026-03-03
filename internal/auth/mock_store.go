package auth

// MockTokenStore is a test double for TokenStore.
// It is placed here (in internal/) to be importable by other internal test packages.
type MockTokenStore struct {
	tokens map[string]string
	err    error
}

// NewMockTokenStore creates a new MockTokenStore.
func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{tokens: make(map[string]string)}
}

// NewMockTokenStoreWithError creates a MockTokenStore that returns the given error.
func NewMockTokenStoreWithError(err error) *MockTokenStore {
	return &MockTokenStore{tokens: make(map[string]string), err: err}
}

// SaveTokens stores tokens in the in-memory map.
func (m *MockTokenStore) SaveTokens(accessToken, refreshToken string) error {
	if m.err != nil {
		return m.err
	}

	m.tokens["access_token"] = accessToken
	m.tokens["refresh_token"] = refreshToken

	return nil
}

// GetAccessToken retrieves the access token from the in-memory map.
func (m *MockTokenStore) GetAccessToken() (string, error) {
	if m.err != nil {
		return "", m.err
	}

	token, ok := m.tokens["access_token"]
	if !ok || token == "" {
		return "", ErrNotAuthenticated
	}

	return token, nil
}

// GetRefreshToken retrieves the refresh token from the in-memory map.
func (m *MockTokenStore) GetRefreshToken() (string, error) {
	if m.err != nil {
		return "", m.err
	}

	token, ok := m.tokens["refresh_token"]
	if !ok || token == "" {
		return "", ErrNotAuthenticated
	}

	return token, nil
}

// ClearTokens removes all stored tokens from the in-memory map.
func (m *MockTokenStore) ClearTokens() error {
	if m.err != nil {
		return m.err
	}

	m.tokens = make(map[string]string)

	return nil
}

// HasValidToken checks whether an access token exists in the in-memory map.
func (m *MockTokenStore) HasValidToken() bool {
	token, ok := m.tokens["access_token"]
	return ok && token != ""
}

// SetTokens is a test helper to pre-populate tokens.
func (m *MockTokenStore) SetTokens(access, refresh string) {
	m.tokens["access_token"] = access
	m.tokens["refresh_token"] = refresh
}
