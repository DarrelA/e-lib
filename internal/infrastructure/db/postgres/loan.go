package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	errMsgExistingLoan         = "User has already borrowed this book."
	errMsgNoActiveLoan         = "No active loan found for this user and book"
	errMsgExistingExtendedLoan = "User has already borrowed and extended this book."
	errMsgBookNotLoan          = "User has not loan this book."
)

type LoanRepository struct {
	dbpool *pgxpool.Pool
}

func NewLoanRepository(dbpool *pgxpool.Pool) postgres.LoanRepository {
	return &LoanRepository{dbpool}
}

var (
	queryCheckExistingLoan = "SELECT COUNT(*) FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"

	execDecrementAvailableCopies = `
	UPDATE books
	SET available_copies = available_copies - 1
	WHERE uuid = $1 AND available_copies > 0
	`

	queryCheckIsExtended = "SELECT is_extended FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"

	queryInsertLoan = `
		INSERT INTO loans (uuid, user_id, book_uuid, name_of_borrower, loan_date, return_date)
    VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW() + interval '4 weeks')
		returning name_of_borrower, loan_date, return_date
	`

	queryExtendReturnDate = `
		UPDATE loans 
		SET return_date = return_date + interval '3 weeks', is_extended = TRUE
		WHERE user_id = $1 
		  AND book_uuid = $2 
		  AND is_returned = FALSE 
		  AND is_extended = FALSE
		returning name_of_borrower, loan_date, return_date
	`

	queryLoanID                  = "SELECT uuid FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"
	execSetIsReturned            = "UPDATE loans SET is_returned = TRUE WHERE uuid = $1 AND is_returned = FALSE"
	execIncrementAvailableCopies = "UPDATE books SET available_copies = available_copies + 1 WHERE uuid = $1 AND available_copies > 0"
)

func (lr LoanRepository) BorrowBook(requestId string, user entity.User, bookDetail *dto.BookDetail) (
	*dto.LoanDetail, *apperrors.RestErr) {

	var existingLoanCount int

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIdKey, requestId)

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
		if errors.Is(err, pgx.ErrNoRows) {
			existingLoanCount = 0
		} else {
			log.Error().Err(err).Msg("")
			return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		}
	}

	if existingLoanCount > 0 {
		return nil, apperrors.NewBadRequestError(errMsgExistingLoan)
	}

	_, err = tx.Exec(ctx, execDecrementAvailableCopies, bookDetail.UUID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	loanDetail := &dto.LoanDetail{BookTitle: bookDetail.Title}
	err = tx.QueryRow(ctx, queryInsertLoan, user.ID, bookDetail.UUID, user.Name).
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

func (lr LoanRepository) ExtendBookLoan(requestId string, user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	loanDetail := &dto.LoanDetail{BookTitle: bookDetail.Title}
	var isExtended bool

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIdKey, requestId)

	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				log.Error().Err(errRollback).Msg("Transaction rollback failed")
			}
		}
	}()

	err = lr.dbpool.QueryRow(ctx, queryCheckIsExtended, user_id, bookDetail.UUID).Scan(&isExtended)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			isExtended = false
		} else {
			log.Error().Err(err).Msg("")
			return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		}
	}

	if isExtended {
		return nil, apperrors.NewBadRequestError(errMsgExistingExtendedLoan)
	}

	err = tx.QueryRow(ctx, queryExtendReturnDate, user_id, bookDetail.UUID).
		Scan(&loanDetail.NameOfBorrower, &loanDetail.LoanDate, &loanDetail.ReturnDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NewNotFoundError(errMsgNoActiveLoan)
		}
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(errMsgBookNotLoan)
	}

	if err = tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	return loanDetail, nil
}

func (lr LoanRepository) ReturnBook(requestId string, user_id int64, book_uuid uuid.UUID) *apperrors.RestErr {
	var loanId uuid.UUID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIdKey, requestId)

	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg("")
			}
		}
	}()

	err = lr.dbpool.QueryRow(ctx, queryLoanID, user_id, book_uuid).Scan(&loanId)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error().Err(err).Msg(errMsgNoActiveLoan)
			return apperrors.NewBadRequestError(errMsgNoActiveLoan)
		}
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(errMsgBookNotFound)
	}

	_, err = lr.dbpool.Exec(ctx, execSetIsReturned, loanId)
	if err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	_, err = lr.dbpool.Exec(ctx, execIncrementAvailableCopies, book_uuid)
	if err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	return nil
}
