package postgres

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
)

type LoanRepository interface {
	BorrowBook(user entity.User, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ExtendBookLoan(user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
	ReturnBook(user_id int64, book_uuid uuid.UUID) *apperrors.RestErr
}
