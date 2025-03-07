package entity

import (
	"time"

	"github.com/google/uuid"
)

type LoanDetail struct {
	UUID           uuid.UUID `json:"uuid"`
	UserID         int64     `json:"user_id"`    // Foreign key to UserDetail
	BookTitle      string    `json:"book_title"` // Foreign key to BookDetail
	NameOfBorrower string    `json:"name_of_borrower"`
	LoanDate       time.Time `json:"loan_date"`
	ReturnDate     time.Time `json:"return_date"`
}
