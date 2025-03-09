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
	var borrowBook dto.BorrowBook
	if err := c.BodyParser(&borrowBook); err != nil {
		log.Error().Err(err).Msg(errMsgInvalidRequestBody)
		return c.Status(fiber.StatusBadRequest).JSON(errMsgInvalidRequestBody)
	}

	loanDetail, err := ls.BorrowBook(borrowBook.Title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) BorrowBook(title string) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, title)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, title))
	}

	loanDetail, err := ls.loanPGDB.BorrowBook(ls.user, bookDetail)
	if err != nil {
		return nil, err
	}

	return loanDetail, nil
}

func (ls *LoanService) ExtendBookLoanHandler(c *fiber.Ctx) error {
	var borrowBook dto.BorrowBook
	if err := c.BodyParser(&borrowBook); err != nil {
		log.Error().Err(err).Msg(errMsgInvalidRequestBody)
		return c.Status(fiber.StatusBadRequest).JSON(errMsgInvalidRequestBody)
	}

	loanDetail, err := ls.ExtendBookLoan(borrowBook.Title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) ExtendBookLoan(title string) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(title)
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
	var borrowBook dto.BorrowBook
	if err := c.BodyParser(&borrowBook); err != nil {
		log.Error().Err(err).Msg(errMsgInvalidRequestBody)
		return c.Status(fiber.StatusBadRequest).JSON(errMsgInvalidRequestBody)
	}

	err := ls.ReturnBook(borrowBook.Title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (ls *LoanService) ReturnBook(title string) *apperrors.RestErr {
	bookDetail, restErr := ls.bookPGDB.GetBook(title)
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
