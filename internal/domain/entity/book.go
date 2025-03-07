package entity

import (
	"time"

	"github.com/google/uuid"
)

type BookDetail struct {
	UUID            *uuid.UUID `json:"uuid"`
	Title           string     `json:"title"`
	AvailableCopies int        `json:"available_copies"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
