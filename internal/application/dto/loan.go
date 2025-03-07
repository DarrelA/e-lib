package dto

import "time"

type LoanDetail struct {
	BookTitle      string    `json:"book_title"`
	NameOfBorrower string    `json:"name_of_borrower"`
	LoanDate       time.Time `json:"loan_date"`
	ReturnDate     time.Time `json:"return_date"`
}
