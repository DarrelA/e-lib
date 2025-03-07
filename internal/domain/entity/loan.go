package entity

import (
	"time"

	"github.com/google/uuid"
)

type Loan struct {
	UUID           uuid.UUID `json:"uuid"`
	UserID         int64     `json:"user_id"`   // Foreign key to UserDetail
	BookUUID       uuid.UUID `json:"book_uuid"` // Foreign key to BookDetail
	NameOfBorrower string    `json:"name_of_borrower"`
	LoanDate       time.Time `json:"loan_date"`
	ReturnDate     time.Time `json:"return_date"`
	IsReturned     bool      `json:"is_returned"`
}
