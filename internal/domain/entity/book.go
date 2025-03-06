package entity

import (
	"time"

	"github.com/google/uuid"
)

type BookDetail struct {
	UUID            *uuid.UUID
	Title           string
	AvailableCopies int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
