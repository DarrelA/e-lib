package postgres

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

type LoanRepository interface {
	BorrowBook(user entity.User, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr)
}
