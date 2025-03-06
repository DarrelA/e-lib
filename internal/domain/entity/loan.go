package entity

import "time"

type LoanDetail struct {
	NameOfBorrower string
	LoanDate       time.Time
	ReturnDate     time.Time
}
