package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	loansJsonFilePath = "./testdata/json/loans.json"

	warnMsgOutOfStock        = "Book '%s' is out of stock."
	errMsgInvalidRequestBody = "Invalid request Body"
)

type LoanService struct {
	user        entity.UserDetail
	bookService appSvc.BookService
}

func NewLoanService(user entity.UserDetail, bookService appSvc.BookService) appSvc.LoanService {
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

func (ls *LoanService) BorrowBook(title string) (*entity.LoanDetail, *apperrors.RestErr) {
	bookDetail, err := ls.bookService.GetBookByTitle(title)
	if err != nil {
		log.Error().Err(err).Msgf("")
		return nil, err
	}

	if bookDetail.AvailableCopies <= 0 {
		log.Warn().Msgf(warnMsgOutOfStock, title)
		return nil, apperrors.NewBadRequestError(fmt.Sprintf(warnMsgOutOfStock, title))
	}

	now := time.Now()
	returnDate := now.Add(time.Hour * 24 * 7 * 4) // Loan for 4 weeks

	loanDetail := &entity.LoanDetail{
		UUID:           uuid.New(),
		UserID:         ls.user.ID, // Use the user from the service context.
		BookID:         *bookDetail.UUID,
		NameOfBorrower: ls.user.Name,
		LoanDate:       now,
		ReturnDate:     returnDate,
	}
	if err := saveLoanDetailToFile(loanDetail); err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	// @TODO: Create a function to decrement available copies.

	return loanDetail, nil
}

func saveLoanDetailToFile(loan *entity.LoanDetail) error {
	filePath := loansJsonFilePath
	existingLoans := []*entity.LoanDetail{}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, &existingLoans)
	if err != nil {
		return err
	}

	existingLoans = append(existingLoans, loan)
	jsonData, err := json.MarshalIndent(existingLoans, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
