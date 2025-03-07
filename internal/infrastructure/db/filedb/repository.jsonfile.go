package filedb

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	booksJsonTestDataPath = "./testdata/json/books.json"

	booksJsonFilePath = "./internal/infrastructure/db/filedb/books.json"
	loansJsonFilePath = "./internal/infrastructure/db/filedb/loans.json"

	errMsgLoanDetailNotFound = "Loan detail not found."
)

var (
	books      []entity.BookDetail
	booksMutex sync.Mutex // Add mutex to synchronize access to `books`
)

func LoadBooksJsonData() []entity.BookDetail {
	jsonData, err := os.ReadFile(booksJsonTestDataPath)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
		return []entity.BookDetail{}
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

func SaveLoanDetail(loan *entity.LoanDetail) error {
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

func DecrementAvailableCopies(title string) error {
	booksMutex.Lock()         // Acquire lock before accessing `books`
	defer booksMutex.Unlock() // Ensure lock is released

	for i := range books {
		if books[i].Title == title {
			if books[i].AvailableCopies > 0 {
				books[i].AvailableCopies--
				break
			}
		}
	}

	if err := saveBooks(books); err != nil {
		return err
	}

	return nil
}

func saveBooks(books []entity.BookDetail) error {
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

func LoadLoanDetails() ([]*entity.LoanDetail, error) {
	jsonData, err := os.ReadFile(loansJsonFilePath)
	if err != nil {
		return nil, err
	}

	var loanDetails []*entity.LoanDetail
	err = json.Unmarshal(jsonData, &loanDetails)

	if err != nil {
		return nil, err
	}

	return loanDetails, nil
}

func UpdateLoanDetail(loanDetails []*entity.LoanDetail, bookDetail *entity.BookDetail, userID int64) (*entity.LoanDetail, error) {
	var updatedLoanDetail *entity.LoanDetail

	found := false
	for i := range loanDetails {
		if loanDetails[i].BookID == *bookDetail.UUID && loanDetails[i].UserID == userID {
			loanDetails[i].ReturnDate = loanDetails[i].ReturnDate.Add(time.Hour * 24 * 7 * 3)
			updatedLoanDetail = loanDetails[i]
			found = true
			break
		}
	}

	if !found {
		log.Error().Msg(errMsgLoanDetailNotFound)
		return nil, apperrors.NewNotFoundError(errMsgLoanDetailNotFound)
	}

	if err := saveLoanDetails(loanDetails); err != nil {
		log.Error().Err(err).Msg("Failed to save loan details")
		return nil, apperrors.NewInternalServerError("Error saving loan details")
	}

	return updatedLoanDetail, nil
}

func saveLoanDetails(loanDetails []*entity.LoanDetail) error {
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
