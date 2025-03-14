package entity

import "time"

type GoogleOAuth2User struct {
	ID            string    `json:"id"`
	UserID        int64     `json:"user_id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	VerifiedEmail bool      `json:"verified_email"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
