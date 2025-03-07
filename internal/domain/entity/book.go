package entity

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	UUID            *uuid.UUID `json:"uuid"`  // Primary Key
	Title           string     `json:"title"` // Unique
	AvailableCopies int        `json:"available_copies"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
