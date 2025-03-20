package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/gofiber/fiber/v2"
)

// LoanService defines the interface for managing loan-related operations (e.g., creating, retrieving, or processing loans).
type LoanService interface {
	BorrowBookHandler(c *fiber.Ctx) error
	BorrowBook(requestID string, userDetail dto.UserDetail, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr)

	ExtendBookLoanHandler(c *fiber.Ctx) error
	ExtendBookLoan(requestID string, userID int64, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr)

	ReturnBookHandler(c *fiber.Ctx) error
	ReturnBook(requestID string, userID int64, bookRequest dto.BookRequest) *apperrors.RestErr
}
