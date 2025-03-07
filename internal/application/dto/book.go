package dto

import "github.com/google/uuid"

type BookDetail struct {
	UUID            uuid.UUID `json:"uuid"`
	Title           string    `json:"title"`
	AvailableCopies int       `json:"available_copies"`
}

type BorrowBook struct {
	Title string `json:"title"`
}
