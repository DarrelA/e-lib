package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

// LoanService defines the interface for managing loan-related operations (e.g., creating, retrieving, or processing loans).
type LoanService interface {
	BorrowBook(title string) (*entity.LoanDetail, *apperrors.RestErr)
	ExtendBookLoan(title string) (*entity.LoanDetail, *apperrors.RestErr)
	ReturnBook(title string) (*entity.LoanDetail, *apperrors.RestErr)
}
