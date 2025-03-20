package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	errMsgNoActiveLoan = "no active loan found for this user and book"
)

type LoanRepository struct {
	dbpool *pgxpool.Pool
}

func NewLoanRepository(dbpool *pgxpool.Pool) repository.LoanRepository {
	return &LoanRepository{dbpool}
}

func (lr LoanRepository) BorrowBook(requestID string, userDetail dto.UserDetail, bookDetail *dto.BookDetail) (
	*dto.LoanDetail, *apperrors.RestErr) {

	var existingLoanCount int

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIDKey, requestID)

	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToBeginTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg(errMsgFailedToRollbackTransaction)
			} else {
				log.Info().Msg(infoMsgRollbackTransactionSuccess)
			}
		} else {
			log.Info().Msg(infoMsgCommittedTransactionSuccess)
		}
	}()

	const queryCheckExistingLoan = "SELECT COUNT(*) FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"
	err = tx.QueryRow(ctx, queryCheckExistingLoan, userDetail.ID, bookDetail.UUID).Scan(&existingLoanCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			existingLoanCount = 0
		} else {
			log.Error().Err(err).Msg("failed to check existing loan")
			rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
			return nil, rErr
		}
	}

	if existingLoanCount > 0 {
		rErr := apperrors.NewBadRequestError("user has already borrowed this book")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	const execDecrementAvailableCopies = "UPDATE books SET available_copies = available_copies - 1 WHERE uuid = $1 AND available_copies > 0"
	_, err = tx.Exec(ctx, execDecrementAvailableCopies, bookDetail.UUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to decrement available copies")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	const queryInsertLoan = `
		INSERT INTO loans (uuid, user_id, book_uuid, name_of_borrower, loan_date, return_date)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW() + interval '4 weeks')
		returning name_of_borrower, loan_date, return_date
	`
	loanDetail := &dto.LoanDetail{BookTitle: bookDetail.Title}
	err = tx.QueryRow(ctx, queryInsertLoan, userDetail.ID, bookDetail.UUID, userDetail.Name).
		Scan(&loanDetail.NameOfBorrower, &loanDetail.LoanDate, &loanDetail.ReturnDate)

	if err != nil {
		log.Error().Err(err).Msg("failed to insert loan")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToCommitTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	err = nil // clear the err, so Rollback wont be executed.

	return loanDetail, nil
}

func (lr LoanRepository) ExtendBookLoan(requestID string, user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	loanDetail := &dto.LoanDetail{BookTitle: bookDetail.Title}
	var isExtended bool

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIDKey, requestID)

	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToBeginTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg(errMsgFailedToRollbackTransaction)
			} else {
				log.Info().Msg(infoMsgRollbackTransactionSuccess)
			}
		} else {
			log.Info().Msg(infoMsgRollbackTransactionSuccess)
		}
	}()

	const queryCheckIsExtended = "SELECT is_extended FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"
	err = tx.QueryRow(ctx, queryCheckIsExtended, user_id, bookDetail.UUID).Scan(&isExtended)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			isExtended = false
		} else {
			log.Error().Err(err).Msg("failed to check if loan is extended")
			rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
			return nil, rErr
		}
	}

	if isExtended {
		rErr := apperrors.NewBadRequestError("user has already borrowed and extended this book")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	const queryExtendReturnDate = `
		UPDATE loans 
		SET return_date = return_date + interval '3 weeks', is_extended = TRUE
		WHERE user_id = $1 
			AND book_uuid = $2 
			AND is_returned = FALSE 
			AND is_extended = FALSE
		returning name_of_borrower, loan_date, return_date
	`
	err = tx.QueryRow(ctx, queryExtendReturnDate, user_id, bookDetail.UUID).
		Scan(&loanDetail.NameOfBorrower, &loanDetail.LoanDate, &loanDetail.ReturnDate)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			rErr := apperrors.NewNotFoundError(errMsgNoActiveLoan)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
			return nil, rErr
		}
		log.Error().Err(err).Msg("failed to extend return date")
		rErr := apperrors.NewInternalServerError("user has not loan this book")
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToCommitTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	err = nil // clear the err, so Rollback wont be executed.

	return loanDetail, nil
}

func (lr LoanRepository) ReturnBook(requestID string, user_id int64, book_uuid uuid.UUID) *apperrors.RestErr {
	var loanID uuid.UUID

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIDKey, requestID)

	tx, err := lr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToBeginTransaction)
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg(errMsgFailedToRollbackTransaction)
			} else {
				log.Info().Msg(infoMsgRollbackTransactionSuccess)
			}
		} else {
			log.Info().Msg(infoMsgRollbackTransactionSuccess)
		}
	}()

	const queryLoanID = "SELECT uuid FROM loans WHERE user_id = $1 AND book_uuid = $2 AND is_returned = FALSE"
	err = tx.QueryRow(ctx, queryLoanID, user_id, book_uuid).Scan(&loanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error().Err(err).Msg(errMsgNoActiveLoan)
			rErr := apperrors.NewBadRequestError(errMsgNoActiveLoan)
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
			return rErr
		}
		log.Error().Err(err).Msg("failed to query loan ID")
		rErr := apperrors.NewInternalServerError(errMsgNoActiveLoan)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return rErr
	}

	const execSetIsReturned = "UPDATE loans SET is_returned = TRUE WHERE uuid = $1 AND is_returned = FALSE"
	_, err = tx.Exec(ctx, execSetIsReturned, loanID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set is_returned")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return rErr
	}

	const execIncrementAvailableCopies = "UPDATE books SET available_copies = available_copies + 1 WHERE uuid = $1 AND available_copies > 0"
	_, err = tx.Exec(ctx, execIncrementAvailableCopies, book_uuid)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment available copies")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return rErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToCommitTransaction)
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	err = nil // clear the err, so Rollback wont be executed.

	return nil
}
