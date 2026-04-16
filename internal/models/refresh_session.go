package models

import "time"

type RefreshSession struct {
	ID            int64      `json:"id" db:"id"`
	UserID        int64      `json:"user_id" db:"user_id"`
	TokenHash     string     `json:"-" db:"token_hash"`
	ExpiresAt     time.Time  `json:"expires_at" db:"expires_at"`
	IssuedAt      time.Time  `json:"issued_at" db:"issued_at"`
	ClientInfo    string     `json:"client_info,omitempty" db:"client_info"`
	ReplacedByID  *int64     `json:"replaced_by_id,omitempty" db:"replaced_by_id"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty" db:"revoked_at"`
	LastAccessJTI string     `json:"-" db:"last_access_jti"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}
