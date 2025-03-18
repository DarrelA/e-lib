package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
)

type LoanRepository interface {
	BorrowBook(requestId string, user entity.User, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ExtendBookLoan(requestId string, user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ReturnBook(requestId string, user_id int64, book_uuid uuid.UUID) *apperrors.RestErr
}
