package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
)

type JsonFileRepository interface {
	LoadBooksJsonData() []entity.Book
	SaveLoanDetail(loan *entity.Loan) error
	DecrementAvailableCopies(title string) error
	IncrementAvailableCopies(title string) error
	LoadLoanDetails() ([]*entity.Loan, error)
	UpdateLoanDetail(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (
		*entity.Loan, *apperrors.RestErr)

	GetLoanStatus(loanDetails []*entity.Loan, loanID uuid.UUID) (bool, bool)
	FindLoanId(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (*uuid.UUID, *apperrors.RestErr)
	SetIsReturned(loanDetails []*entity.Loan, loanID uuid.UUID) *apperrors.RestErr
}
