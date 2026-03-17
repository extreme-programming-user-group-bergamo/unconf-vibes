package models

import "time"

type User struct {
	ID             int64     `json:"id" db:"id"`
	GitHubID       string    `json:"github_id" db:"github_id"`
	Email          string    `json:"email" db:"email"`
	DisplayName    string    `json:"display_name" db:"display_name"`
	PrivacySetting string    `json:"privacy_setting" db:"privacy_setting"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
