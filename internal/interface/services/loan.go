package services

import (
	"fmt"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	warnMsgOutOfStock = "book '%s' is out of stock"
)

type LoanService struct {
	bookPGDB repository.BookRepository
	loanPGDB repository.LoanRepository
}

func NewLoanService(bookPGDB repository.BookRepository, loanPGDB repository.LoanRepository) appSvc.LoanService {
	return &LoanService{bookPGDB, loanPGDB}
}

func (ls *LoanService) BorrowBookHandler(c *fiber.Ctx) error {
	borrowBook, requestID, userDetail, restErr := getContextInfo(c)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	loanDetail, err := ls.BorrowBook(requestID, userDetail, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) BorrowBook(requestID string, userDetail dto.UserDetail, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestID, bookRequest.Title)
	if restErr != nil {
		log.Warn().Msg(errMsgBookNotFound + ":" + restErr.Message)
		restErr.Message = errMsgBookNotFound
		return nil, restErr
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, bookRequest)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, bookRequest))
	}

	loanDetail, err := ls.loanPGDB.BorrowBook(requestID, userDetail, bookDetail)
	if err != nil {
		return nil, err
	}

	return loanDetail, nil
}

func (ls *LoanService) ExtendBookLoanHandler(c *fiber.Ctx) error {
	borrowBook, requestID, userDetail, restErr := getContextInfo(c)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	loanDetail, err := ls.ExtendBookLoan(requestID, userDetail.ID, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (ls *LoanService) ExtendBookLoan(requestID string, userID int64, bookRequest dto.BookRequest) (*dto.LoanDetail, *apperrors.RestErr) {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestID, bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	loanDetail, restErr := ls.loanPGDB.ExtendBookLoan(requestID, userID, bookDetail)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	return loanDetail, nil
}

func (ls *LoanService) ReturnBookHandler(c *fiber.Ctx) error {
	borrowBook, requestID, userDetail, restErr := getContextInfo(c)
	if restErr != nil {
		return c.Status(restErr.Status).JSON(restErr)
	}

	err := ls.ReturnBook(requestID, userDetail.ID, borrowBook)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (ls *LoanService) ReturnBook(requestID string, userID int64, bookRequest dto.BookRequest) *apperrors.RestErr {
	bookDetail, restErr := ls.bookPGDB.GetBook(requestID, bookRequest.Title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	restErr = ls.loanPGDB.ReturnBook(requestID, userID, bookDetail.UUID)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	return nil
}

func getContextInfo(c *fiber.Ctx) (dto.BookRequest, string, dto.UserDetail, *apperrors.RestErr) {
	restErr := apperrors.NewBadRequestError(apperrors.ErrMsgSomethingWentWrong)
	borrowBook, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		log.Error().Msgf(errMsgNotFoundOrIncorrectType, "bookTitleKey")
		return dto.BookRequest{}, "", dto.UserDetail{}, restErr
	}

	requestID, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msgf(errMsgNotFoundOrIncorrectType, "requestid")
		return dto.BookRequest{}, "", dto.UserDetail{}, restErr
	}

	userDetail, ok := c.Locals("userDetail").(dto.UserDetail)
	if !ok {
		log.Error().Msgf(errMsgNotFoundOrIncorrectType, "userDetail")
		return dto.BookRequest{}, "", dto.UserDetail{}, restErr
	}

	return borrowBook, requestID, userDetail, nil
}
