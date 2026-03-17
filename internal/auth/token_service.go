package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	paseto "aidanwoods.dev/go-paseto"
)

type AccessTokenInput struct {
	UserID     int64
	SessionID  int64
	Issuer     string
	Audience   string
	NotBefore  time.Time
	IssuedAt   time.Time
	JTI        string
	ExpiryTime time.Time
}

type AccessTokenClaims struct {
	UserID    int64
	SessionID int64
	JTI       string
	IssuedAt  time.Time
	NotBefore time.Time
	ExpiresAt time.Time
}

type TokenService struct {
	v4      paseto.V4SymmetricKey
	nowFunc func() time.Time
}

func NewTokenService(symmetricKey string) (*TokenService, error) {
	if len(symmetricKey) != 32 {
		return nil, fmt.Errorf("failed to initialize token service: paseto symmetric key must be exactly 32 bytes")
	}

	key, err := paseto.V4SymmetricKeyFromBytes([]byte(symmetricKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize token service symmetric key: %w", err)
	}

	return &TokenService{v4: key, nowFunc: time.Now}, nil
}

func (s *TokenService) IssueAccessToken(ctx context.Context, input AccessTokenInput) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("failed to issue access token: %w", err)
	}

	now := s.nowFunc().UTC()
	issuedAt := input.IssuedAt.UTC()
	if issuedAt.IsZero() {
		issuedAt = now
	}

	notBefore := input.NotBefore.UTC()
	if notBefore.IsZero() {
		notBefore = now
	}

	expiresAt := input.ExpiryTime.UTC()
	if expiresAt.IsZero() || !expiresAt.After(now) {
		return "", fmt.Errorf("failed to issue access token: expiry time must be in the future")
	}

	token := paseto.NewToken()
	token.SetIssuer(input.Issuer)
	token.SetAudience(input.Audience)
	token.SetSubject(strconv.FormatInt(input.UserID, 10))
	token.SetIssuedAt(issuedAt)
	token.SetNotBefore(notBefore)
	token.SetExpiration(expiresAt)
	token.SetJti(input.JTI)
	token.SetString("sid", strconv.FormatInt(input.SessionID, 10))

	encrypted := token.V4Encrypt(s.v4, nil)
	return encrypted, nil
}

func (s *TokenService) GenerateTokenID() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("failed to generate token id: %w", err)
	}

	return hex.EncodeToString(buffer), nil
}

func (s *TokenService) GenerateRefreshToken() (string, string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	raw := base64.RawURLEncoding.EncodeToString(buffer)
	hash := hashRefreshToken(raw)
	return raw, hash, nil
}

const (
	clockSkewTolerance = 60 * time.Second
	maxTokenAge        = 24 * time.Hour
)

// ValidateToken decrypts and validates a PASETO v4.local token, returning the embedded claims.
func (s *TokenService) ValidateToken(ctx context.Context, encrypted string) (*AccessTokenClaims, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	parser := paseto.NewParserWithoutExpiryCheck()
	token, err := parser.ParseV4Local(s.v4, encrypted, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	iss, err := token.GetIssuer()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: missing issuer claim")
	}
	if iss != "unconf-api" {
		return nil, fmt.Errorf("failed to validate token: invalid issuer %q", iss)
	}

	aud, err := token.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: missing audience claim")
	}
	if aud != "unconf-cli" {
		return nil, fmt.Errorf("failed to validate token: invalid audience %q", aud)
	}

	now := s.nowFunc().UTC()

	exp, err := token.GetExpiration()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: missing expiration claim")
	}
	if now.After(exp.Add(clockSkewTolerance)) {
		return nil, fmt.Errorf("failed to validate token: token expired")
	}

	nbf, err := token.GetNotBefore()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: missing not-before claim")
	}
	if now.Add(clockSkewTolerance).Before(nbf) {
		return nil, fmt.Errorf("failed to validate token: token not yet valid")
	}

	iat, err := token.GetIssuedAt()
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: missing issued-at claim")
	}
	if now.Sub(iat) > maxTokenAge+clockSkewTolerance {
		return nil, fmt.Errorf("failed to validate token: token exceeds maximum age")
	}

	jti, err := token.GetJti()
	if err != nil || jti == "" {
		return nil, fmt.Errorf("failed to validate token: missing jti claim")
	}

	sub, err := token.GetSubject()
	if err != nil || sub == "" {
		return nil, fmt.Errorf("failed to validate token: missing subject claim")
	}
	userID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: invalid subject: %w", err)
	}

	var sidStr string
	if err := token.Get("sid", &sidStr); err != nil || sidStr == "" {
		return nil, fmt.Errorf("failed to validate token: missing session id claim")
	}
	sessionID, err := strconv.ParseInt(sidStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: invalid session id: %w", err)
	}

	return &AccessTokenClaims{
		UserID:    userID,
		SessionID: sessionID,
		JTI:       jti,
		IssuedAt:  iat,
		NotBefore: nbf,
		ExpiresAt: exp,
	}, nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func HashRefreshToken(token string) string {
	return hashRefreshToken(token)
}
