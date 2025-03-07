package services

import (
	"fmt"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/filedb"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	warnMsgOutOfStock        = "Book '%s' is out of stock."
	errMsgInvalidRequestBody = "Invalid request Body"
	errMsgOnlyOneCopy        = "Each user can only borrow one copy per book."
)

type LoanService struct {
	user        entity.User
	bookService appSvc.BookService
}

func NewLoanService(user entity.User, bookService appSvc.BookService) appSvc.LoanService {
	return &LoanService{user: user, bookService: bookService}
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
	bookDetail, restErr := ls.bookService.GetBookByTitle(title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	loanDetails, err := filedb.LoadLoanDetails()
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	hasLoanId := filedb.HasLoanId(loanDetails, bookDetail, ls.user.ID)
	if hasLoanId {
		return nil, apperrors.NewBadRequestError(errMsgOnlyOneCopy)
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, title)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, title))
	}

	now := time.Now()
	returnDate := now.Add(time.Hour * 24 * 7 * 4) // Loan for 4 weeks

	newLoan := &entity.Loan{
		UUID:           uuid.New(),
		UserID:         ls.user.ID, // Use the user from the service context.
		BookUUID:       bookDetail.UUID,
		NameOfBorrower: ls.user.Name,
		LoanDate:       now,
		ReturnDate:     returnDate,
		IsReturned:     false,
	}
	if err := filedb.SaveLoanDetail(newLoan); err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	if err := filedb.DecrementAvailableCopies(title); err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	loanDetail := &dto.LoanDetail{
		BookTitle:      title,
		NameOfBorrower: newLoan.NameOfBorrower,
		LoanDate:       newLoan.LoanDate,
		ReturnDate:     newLoan.ReturnDate,
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
	loanDetails, err := filedb.LoadLoanDetails()
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	bookDetail, restErr := ls.bookService.GetBookByTitle(title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	updatedLoanDetail, restErr := filedb.UpdateLoanDetail(loanDetails, bookDetail, ls.user.ID)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return nil, restErr
	}

	loanDetail := &dto.LoanDetail{
		BookTitle:      title,
		NameOfBorrower: updatedLoanDetail.NameOfBorrower,
		LoanDate:       updatedLoanDetail.LoanDate,
		ReturnDate:     updatedLoanDetail.ReturnDate,
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
	if err := filedb.IncrementAvailableCopies(title); err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	bookDetail, restErr := ls.bookService.GetBookByTitle(title)
	if restErr != nil {
		log.Error().Err(restErr).Msgf("")
		return restErr
	}

	loanDetails, err := filedb.LoadLoanDetails()
	if err != nil {
		log.Error().Err(err).Msg("")
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	loanDetail, restErr := filedb.FindLoanId(loanDetails, bookDetail, ls.user.ID)
	if restErr != nil {
		log.Error().Err(restErr).Msg("")
		return restErr
	}

	if restErr := filedb.SetIsReturned(loanDetails, loanDetail.UUID); restErr != nil {
		log.Error().Err(err).Msg("")
		return restErr
	}

	return nil
}
