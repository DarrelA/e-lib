package services

import (
	"fmt"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
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
	borrowBook, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		log.Error().Msg("bookTitleKey not found or has incorrect type.")
	}

	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msg("requestid not found or has incorrect type.")
	}

	loanDetail, err := ls.BorrowBook(requestId, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) BorrowBook(requestId string, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestId, bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, bookRequest)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, bookRequest))
	}

	loanDetail, err := ls.loanPGDB.BorrowBook(requestId, ls.user, bookDetail)
	if err != nil {
		return nil, err
	}

	return loanDetail, nil
}

func (ls *LoanService) ExtendBookLoanHandler(c *fiber.Ctx) error {
	borrowBook, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		log.Error().Msg("bookTitleKey not found or has incorrect type.")
	}

	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msg("requestid not found or has incorrect type.")
	}

	loanDetail, err := ls.ExtendBookLoan(requestId, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) ExtendBookLoan(requestId string, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestId, bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	loanDetail, restErr := ls.loanPGDB.ExtendBookLoan(requestId, ls.user.ID, bookDetail)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	return loanDetail, nil
}

func (ls *LoanService) ReturnBookHandler(c *fiber.Ctx) error {
	borrowBook, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		log.Error().Msg("bookTitleKey not found or has incorrect type.")
	}

	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msg("requestid not found or has incorrect type.")
	}

	err := ls.ReturnBook(requestId, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (ls *LoanService) ReturnBook(requestId string, bookRequest dto.BookRequest) *apperrors.RestErr {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestId, bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	restErr = ls.loanPGDB.ReturnBook(requestId, ls.user.ID, bookDetail.UUID)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	return nil
}
