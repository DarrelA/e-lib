package services

import (
	"fmt"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	repository "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	warnMsgOutOfStock        = "Book '%s' is out of stock."
	errMsgInvalidRequestBody = "Invalid request Body"
	errMsgOnlyOneCopy        = "Each user can only borrow one copy per book."
	errMsgAlreadyReturned    = "You have already returned the book."
)

type LoanService struct {
	user     entity.User
	bookPGDB repository.BookRepository
	loanPGDB repository.LoanRepository
}

func NewLoanService(
	user entity.User, bookPGDB repository.BookRepository, loanPGDB repository.LoanRepository) appSvc.LoanService {
	return &LoanService{user, bookPGDB, loanPGDB}
}

func (ls *LoanService) BorrowBookHandler(c *fiber.Ctx) error {
	borrowBook := c.Locals("bookTitleKey").(dto.BookRequest)
	loanDetail, err := ls.BorrowBook(borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) BorrowBook(bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, bookRequest)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, bookRequest))
	}

	loanDetail, err := ls.loanPGDB.BorrowBook(ls.user, bookDetail)
	if err != nil {
		return nil, err
	}

	return loanDetail, nil
}

func (ls *LoanService) ExtendBookLoanHandler(c *fiber.Ctx) error {
	borrowBook := c.Locals("bookTitleKey").(dto.BookRequest)
	loanDetail, err := ls.ExtendBookLoan(borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) ExtendBookLoan(bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	loanDetail, restErr := ls.loanPGDB.ExtendBookLoan(ls.user.ID, bookDetail)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	return loanDetail, nil
}

func (ls *LoanService) ReturnBookHandler(c *fiber.Ctx) error {
	borrowBook := c.Locals("bookTitleKey").(dto.BookRequest)
	err := ls.ReturnBook(borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (ls *LoanService) ReturnBook(bookRequest dto.BookRequest) *apperrors.RestErr {
	bookDetail, restErr := ls.bookPGDB.GetBook(bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	restErr = ls.loanPGDB.ReturnBook(ls.user.ID, bookDetail.UUID)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	return nil
}
