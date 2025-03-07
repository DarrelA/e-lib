package filedb

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	booksJsonTestDataPath = "./testdata/json/books.json"

	booksJsonFilePath = "./internal/infrastructure/db/filedb/books.json"
	loansJsonFilePath = "./internal/infrastructure/db/filedb/loans.json"

	errMsgLoanDetailNotFound = "Loan detail not found."
	errMsgFailedToSaveLoan   = "Failed to save loan details."
)

var (
	books      []entity.Book
	booksMutex sync.Mutex // Add mutex to synchronize access to `books`
)

func LoadBooksJsonData() []entity.Book {
	jsonData, err := os.ReadFile(booksJsonTestDataPath)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
		return []entity.Book{}
	}

	err = json.Unmarshal(jsonData, &books)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	for i := range books {
		newUUID := uuid.New()
		books[i].UUID = &newUUID

		currentTime := time.Now()
		books[i].CreatedAt = currentTime
		books[i].UpdatedAt = currentTime
	}

	return books
}

func SaveLoanDetail(loan *entity.Loan) error {
	filePath := loansJsonFilePath
	existingLoans := []*entity.Loan{}
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

func DecrementAvailableCopies(title string) error {
	booksMutex.Lock()         // Acquire lock before accessing `books`
	defer booksMutex.Unlock() // Ensure lock is released

	for _, book := range books {
		if book.Title == title {
			if book.AvailableCopies > 0 {
				book.AvailableCopies--
				break
			}
		}
	}

	if err := saveBooks(books); err != nil {
		return err
	}

	return nil
}

func IncrementAvailableCopies(title string) error {
	booksMutex.Lock()         // Acquire lock before accessing `books`
	defer booksMutex.Unlock() // Ensure lock is released

	for _, book := range books {
		if book.Title == title {
			book.AvailableCopies++
			break
		}
	}

	if err := saveBooks(books); err != nil {
		return err
	}

	return nil
}

func saveBooks(books []entity.Book) error {
	jsonData, err := json.MarshalIndent(books, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(booksJsonFilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadLoanDetails() ([]*entity.Loan, error) {
	jsonData, err := os.ReadFile(loansJsonFilePath)
	if err != nil {
		return nil, err
	}

	var loanDetails []*entity.Loan
	err = json.Unmarshal(jsonData, &loanDetails)

	if err != nil {
		return nil, err
	}

	return loanDetails, nil
}

func UpdateLoanDetail(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (
	*entity.Loan, *apperrors.RestErr) {
	var updatedLoanDetail *entity.Loan

	found := false
	for _, loan := range loanDetails {
		if loan.BookUUID == bookDetail.UUID && loan.UserID == userID {
			loan.ReturnDate = loan.ReturnDate.Add(time.Hour * 24 * 7 * 3)
			updatedLoanDetail = loan
			found = true
			break
		}
	}

	if !found {
		log.Error().Msg(errMsgLoanDetailNotFound)
		return nil, apperrors.NewNotFoundError(errMsgLoanDetailNotFound)
	}

	if err := saveLoanDetails(loanDetails); err != nil {
		log.Error().Err(err).Msg(errMsgFailedToSaveLoan)
		return nil, apperrors.NewInternalServerError(errMsgFailedToSaveLoan)
	}

	return updatedLoanDetail, nil
}

func HasLoanId(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) bool {
	for _, loan := range loanDetails {
		if loan.BookUUID == bookDetail.UUID && loan.UserID == userID {
			return true
		}
	}
	return false
}

func FindLoanId(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (*entity.Loan, *apperrors.RestErr) {
	var loanDetail *entity.Loan
	found := false
	for _, loan := range loanDetails {
		if loan.BookUUID == bookDetail.UUID && loan.UserID == userID {
			loanDetail = loan
			found = true
			break
		}
	}

	if !found {
		log.Error().Msg(errMsgLoanDetailNotFound)
		return nil, apperrors.NewNotFoundError(errMsgLoanDetailNotFound)
	}

	return loanDetail, nil
}

func SetIsReturned(loanDetails []*entity.Loan, loanID uuid.UUID) *apperrors.RestErr {
	found := false
	for _, loan := range loanDetails {
		if loan.UUID == loanID {
			loan.IsReturned = true
			found = true
			break
		}
	}

	if !found {
		log.Error().Msg(errMsgLoanDetailNotFound)
		return apperrors.NewNotFoundError(errMsgLoanDetailNotFound)
	}

	if err := saveLoanDetails(loanDetails); err != nil {
		log.Error().Err(err).Msg(errMsgFailedToSaveLoan)
		return apperrors.NewInternalServerError(errMsgFailedToSaveLoan)
	}

	return nil
}

func saveLoanDetails(loanDetails []*entity.Loan) error {
	jsonData, err := json.MarshalIndent(loanDetails, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(loansJsonFilePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
