package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSymmetricKey = "0123456789abcdef0123456789abcdef"

func newTestTokenService(t *testing.T) *TokenService {
	t.Helper()

	ts, err := NewTokenService(testSymmetricKey)
	require.NoError(t, err)

	return ts
}

func issueTestToken(t *testing.T, ts *TokenService, opts AccessTokenInput) string {
	t.Helper()

	token, err := ts.IssueAccessToken(context.Background(), opts)
	require.NoError(t, err)

	return token
}

func defaultTokenInput(now time.Time) AccessTokenInput {
	return AccessTokenInput{
		UserID:     42,
		SessionID:  99,
		Issuer:     "unconf-api",
		Audience:   "unconf-cli",
		NotBefore:  now,
		IssuedAt:   now,
		JTI:        "test-jti-123",
		ExpiryTime: now.Add(1 * time.Hour),
	}
}

func TestValidateToken_RoundTrip(t *testing.T) {
	ts := newTestTokenService(t)
	now := time.Now().UTC()
	ts.nowFunc = func() time.Time { return now }

	encrypted := issueTestToken(t, ts, defaultTokenInput(now))

	claims, err := ts.ValidateToken(context.Background(), encrypted)
	require.NoError(t, err)

	assert.Equal(t, int64(42), claims.UserID)
	assert.Equal(t, int64(99), claims.SessionID)
	assert.Equal(t, "test-jti-123", claims.JTI)
	assert.WithinDuration(t, now, claims.IssuedAt, time.Second)
	assert.WithinDuration(t, now, claims.NotBefore, time.Second)
	assert.WithinDuration(t, now.Add(1*time.Hour), claims.ExpiresAt, time.Second)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	ts := newTestTokenService(t)
	issuedAt := time.Now().UTC()
	ts.nowFunc = func() time.Time { return issuedAt }

	encrypted := issueTestToken(t, ts, defaultTokenInput(issuedAt))

	// Move time past expiry + clock skew
	ts.nowFunc = func() time.Time { return issuedAt.Add(2 * time.Hour) }

	_, err := ts.ValidateToken(context.Background(), encrypted)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token expired")
}

func TestValidateToken_WrongIssuer(t *testing.T) {
	ts := newTestTokenService(t)
	now := time.Now().UTC()
	ts.nowFunc = func() time.Time { return now }

	input := defaultTokenInput(now)
	input.Issuer = "wrong-issuer"
	encrypted := issueTestToken(t, ts, input)

	_, err := ts.ValidateToken(context.Background(), encrypted)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid issuer")
}

func TestValidateToken_WrongAudience(t *testing.T) {
	ts := newTestTokenService(t)
	now := time.Now().UTC()
	ts.nowFunc = func() time.Time { return now }

	input := defaultTokenInput(now)
	input.Audience = "wrong-audience"
	encrypted := issueTestToken(t, ts, input)

	_, err := ts.ValidateToken(context.Background(), encrypted)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid audience")
}

func TestValidateToken_FutureNotBefore(t *testing.T) {
	ts := newTestTokenService(t)
	now := time.Now().UTC()
	ts.nowFunc = func() time.Time { return now }

	input := defaultTokenInput(now)
	input.NotBefore = now.Add(5 * time.Minute) // Beyond clock skew
	encrypted := issueTestToken(t, ts, input)

	_, err := ts.ValidateToken(context.Background(), encrypted)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token not yet valid")
}

func TestValidateToken_TokenAgeExceeds24Hours(t *testing.T) {
	ts := newTestTokenService(t)
	issuedAt := time.Now().UTC()
	ts.nowFunc = func() time.Time { return issuedAt }

	input := defaultTokenInput(issuedAt)
	input.ExpiryTime = issuedAt.Add(48 * time.Hour) // Far future expiry
	encrypted := issueTestToken(t, ts, input)

	// Move time forward 25 hours (past 24h max age + clock skew)
	ts.nowFunc = func() time.Time { return issuedAt.Add(25 * time.Hour) }

	_, err := ts.ValidateToken(context.Background(), encrypted)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token exceeds maximum age")
}

func TestValidateToken_CorruptedToken(t *testing.T) {
	ts := newTestTokenService(t)

	_, err := ts.ValidateToken(context.Background(), "garbage-token-data")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate token")
}

func TestValidateToken_CancelledContext(t *testing.T) {
	ts := newTestTokenService(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ts.ValidateToken(ctx, "any-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate token")
}
