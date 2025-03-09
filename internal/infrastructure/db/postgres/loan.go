package postgres

import (
	"context"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	errMsgExistingLoan = "User has already borrowed this book."
)

type LoanRepository struct {
	dbpool *pgxpool.Pool
}

func NewLoanRepository(dbpool *pgxpool.Pool) postgres.LoanRepository {
	return &LoanRepository{dbpool}
}

var (
	queryCheckExistingLoan = "SELECT COUNT(*) FROM Loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"

	queryDecrementAvailableCopies = `
		UPDATE books
		SET available_copies = available_copies - 1
		WHERE uuid = $1 AND available_copies > 0
	`

	insertLoan = `
		INSERT INTO Loans (uuid, user_id, book_uuid, name_of_borrower, loan_date, return_date)
    VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW() + interval '4 weeks')
	`
)

func (lr LoanRepository) BorrowBook(user entity.User, bookDetail *dto.BookDetail) (
	*dto.LoanDetail, *apperrors.RestErr) {

	var existingLoanCount int
	ctx := context.Background()
	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg("")
			}
		}
	}()

	err = lr.dbpool.QueryRow(ctx, queryCheckExistingLoan, user.ID, bookDetail.UUID).Scan(&existingLoanCount)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	if existingLoanCount > 0 {
		return nil, apperrors.NewBadRequestError(errMsgExistingLoan)
	}

	_, err = tx.Exec(ctx, queryDecrementAvailableCopies, bookDetail.UUID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	loanDetail := &dto.LoanDetail{BookTitle: bookDetail.Title}
	err = lr.dbpool.QueryRow(ctx, insertLoan+
		" returning name_of_borrower, loan_date, return_date",
		user.ID, bookDetail.UUID, user.Name).
		Scan(&loanDetail.NameOfBorrower, &loanDetail.LoanDate, &loanDetail.ReturnDate)

	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	return loanDetail, nil
}
