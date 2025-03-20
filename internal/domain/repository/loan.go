package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/google/uuid"
)

type LoanRepository interface {
	BorrowBook(requestID string, userDetail dto.UserDetail, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ExtendBookLoan(requestID string, user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ReturnBook(requestID string, user_id int64, book_uuid uuid.UUID) *apperrors.RestErr
}
