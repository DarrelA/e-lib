package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
)

// LoanService defines the interface for managing loan-related operations (e.g., creating, retrieving, or processing loans).
type LoanService interface {
	BorrowBookHandler(c *fiber.Ctx) error
	BorrowBook(title string) (*entity.LoanDetail, *apperrors.RestErr)

	ExtendBookLoanHandler(c *fiber.Ctx) error
	ExtendBookLoan(title string) (*entity.LoanDetail, *apperrors.RestErr)
}
