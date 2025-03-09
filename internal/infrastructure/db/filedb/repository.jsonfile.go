package filedb

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository/filedb"
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

type JsonFileService struct{}

func NewJsonFileService() filedb.JsonFileRepository {
	return &JsonFileService{}
}

func (js JsonFileService) LoadBooksJsonData() []entity.Book {
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

func (js JsonFileService) SaveLoanDetail(loan *entity.Loan) error {
	filePath := loansJsonFilePath
	existingLoans := []*entity.Loan{}
	if err := readJSONFromFile(filePath, &existingLoans); err != nil {
		if os.IsNotExist(err) {
			existingLoans = []*entity.Loan{}
		} else {
			return err
		}
	}

	existingLoans = append(existingLoans, loan)
	return writeJSONToFile(filePath, existingLoans)
}

func (js JsonFileService) DecrementAvailableCopies(title string) error {
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

func (js JsonFileService) IncrementAvailableCopies(title string) error {
	booksMutex.Lock()         // Acquire lock before accessing `books`
	defer booksMutex.Unlock() // Ensure lock is released

	for i := range books {
		if books[i].Title == title {
			books[i].AvailableCopies++
			break
		}
	}

	if err := saveBooks(books); err != nil {
		return err
	}

	return nil
}

func saveBooks(books []entity.Book) error {
	return writeJSONToFile(booksJsonFilePath, books)
}

func (js JsonFileService) LoadLoanDetails() ([]*entity.Loan, error) {
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

func (js JsonFileService) UpdateLoanDetail(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (
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

func (js JsonFileService) GetLoanStatus(loanDetails []*entity.Loan, loanID uuid.UUID) (bool, bool) {
	for _, loan := range loanDetails {
		if loan.UUID == loanID {
			return true, loan.IsReturned
		}
	}
	return false, false
}

func (js JsonFileService) FindLoanId(loanDetails []*entity.Loan, bookDetail *dto.BookDetail, userID int64) (*uuid.UUID, *apperrors.RestErr) {
	var loanID *uuid.UUID
	found := false
	for _, loan := range loanDetails {
		if loan.BookUUID == bookDetail.UUID && loan.UserID == userID && !loan.IsReturned {
			loanID = &loan.UUID
			found = true
			break
		}
	}

	if !found {
		log.Error().Msg(errMsgLoanDetailNotFound)
		return nil, apperrors.NewNotFoundError(errMsgLoanDetailNotFound)
	}

	return loanID, nil
}

func (js JsonFileService) SetIsReturned(loanDetails []*entity.Loan, loanID uuid.UUID) *apperrors.RestErr {
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
	return writeJSONToFile(loansJsonFilePath, loanDetails)
}

// Helper function to read JSON data from a file and unmarshal it.
func readJSONFromFile(filePath string, data interface{}) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, data)
}

// Helper function to marshal data to JSON and write it to a file.
func writeJSONToFile(filePath string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0644)
}
