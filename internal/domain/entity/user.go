package entity

import (
	"time"
)

type UserDetail struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MyInfoResponse struct {
	Name struct {
		Value string `json:"value"`
	} `json:"name"`

	Email struct {
		Value string `json:"value"`
	} `json:"email"`
}
