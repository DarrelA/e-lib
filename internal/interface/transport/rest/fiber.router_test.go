package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	bookTitle = "Anna"
)

type mockBookRepository struct {
	mock.Mock
	ExpectedBook *dto.BookDetail
}

func (m *mockBookRepository) GetBook(title string) (*dto.BookDetail, *apperrors.RestErr) {
	args := m.Called(title)
	book, ok := args.Get(0).(*dto.BookDetail)
	if !ok {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return book, nil
}

type mockLoanRepository struct{ mock.Mock }

func (m *mockLoanRepository) BorrowBook(user entity.User, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	args := m.Called(user, bookDetail.UUID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*dto.LoanDetail), nil
}

func (m *mockLoanRepository) ExtendBookLoan(user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	args := m.Called(user_id, bookDetail)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*dto.LoanDetail), nil
}

func (m *mockLoanRepository) ReturnBook(user_id int64, book_uuid uuid.UUID) *apperrors.RestErr {
	args := m.Called(user_id, book_uuid)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*apperrors.RestErr)
}

func TestRoutes(t *testing.T) {
	// Shared setup
	testUser := entity.User{ID: 1, Name: "User1"}
	bookUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	expectedBook := dto.BookDetail{UUID: bookUUID, Title: "anna", AvailableCopies: 10}

	now := time.Now().UTC()
	expectedLoan := dto.LoanDetail{
		BookTitle:      "anna",
		NameOfBorrower: "User1",
		LoanDate:       now,
		ReturnDate:     now.Add(time.Hour * 24 * 7 * 4),
	}

	extendedReturnDate := now.Add(time.Hour * 24 * 7 * 3)
	extendedLoan := dto.LoanDetail{
		BookTitle:      "anna",
		NameOfBorrower: "User1",
		LoanDate:       now,
		ReturnDate:     extendedReturnDate,
	}

	mockBookRepo := new(mockBookRepository)
	mockBookRepo.On("GetBook", "anna").Return(&expectedBook, nil)

	bookService := interfaceSvc.NewBookService(mockBookRepo)

	mockLoanRepo := new(mockLoanRepository)
	mockLoanRepo.On("BorrowBook", testUser, bookUUID).Return(&expectedLoan, nil)
	mockLoanRepo.On("ExtendBookLoan", testUser.ID, &expectedBook).Return(&extendedLoan, nil)
	mockLoanRepo.On("ReturnBook", testUser.ID, bookUUID).Return(nil, nil)

	loanService := interfaceSvc.NewLoanService(testUser, mockBookRepo, mockLoanRepo)
	app := NewRouter(bookService, loanService)

	t.Run("GetBookByTitle", func(t *testing.T) {
		tests := []struct {
			description  string
			route        string
			expectedCode int
			expectedBody string
		}{
			{
				description:  "Get existing book by title",
				route:        fmt.Sprintf("/Book?title=%s", bookTitle),
				expectedCode: http.StatusOK,
				expectedBody: `{"uuid":"123e4567-e89b-12d3-a456-426614174000","title":"anna","available_copies":10}`,
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				req := httptest.NewRequest("GET", test.route, nil)
				resp, err := app.Test(req)
				assert.Nil(t, err)
				assert.Equal(t, test.expectedCode, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.JSONEq(t, test.expectedBody, string(body))
			})
		}
	})

	t.Run("BorrowBook", func(t *testing.T) {
		tests := []struct {
			description  string
			route        string
			method       string
			requestBody  dto.BorrowBook
			expectedCode int
			expectedBody dto.LoanDetail
		}{
			{
				description:  "Successfully borrow book for 4 weeks",
				route:        "/Borrow",
				method:       http.MethodPost,
				requestBody:  dto.BorrowBook{Title: bookTitle},
				expectedCode: http.StatusOK,
				expectedBody: expectedLoan,
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				reqBody, _ := json.Marshal(test.requestBody)
				req := httptest.NewRequest(test.method, test.route, bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				assert.Nil(t, err)
				assert.Equal(t, test.expectedCode, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var actualLoan dto.LoanDetail
				err = json.Unmarshal(body, &actualLoan)
				assert.Nil(t, err)

				// Check non-date fields
				assert.Equal(t, test.expectedBody.BookTitle, actualLoan.BookTitle, "Book title mismatch")
				assert.Equal(t, test.expectedBody.NameOfBorrower, actualLoan.NameOfBorrower, "Borrower name mismatch")

				// Check LoanDate is recent (within 5 seconds of current UTC time)
				nowUTC := time.Now().UTC()
				loanDateUTC := actualLoan.LoanDate.UTC()
				assert.WithinDuration(t, nowUTC, loanDateUTC, 5*time.Second, "LoanDate should be within 5 seconds of now")

				// Check ReturnDate is exactly 4 weeks after LoanDate
				expectedReturnDate := actualLoan.LoanDate.Add(time.Hour * 24 * 7 * 4)
				assert.True(t, expectedReturnDate.Equal(actualLoan.ReturnDate),
					"ReturnDate should be 4 weeks after LoanDate (expected %s, got %s)", expectedReturnDate, actualLoan.ReturnDate)
			})
		}
	})

	t.Run("ExtendBook", func(t *testing.T) {
		tests := []struct {
			description  string
			route        string
			method       string
			requestBody  dto.BorrowBook
			expectedCode int
			expectedBody dto.LoanDetail
		}{
			{
				description:  "Successfully extend the loan of the book by 3 weeks",
				route:        "/Extend",
				method:       http.MethodPost,
				requestBody:  dto.BorrowBook{Title: bookTitle},
				expectedCode: http.StatusOK,
				expectedBody: expectedLoan,
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				reqBody, _ := json.Marshal(test.requestBody)
				req := httptest.NewRequest(test.method, test.route, bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				assert.Nil(t, err)
				assert.Equal(t, test.expectedCode, resp.StatusCode)

				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()

				var actualLoan dto.LoanDetail
				err = json.Unmarshal(body, &actualLoan)
				assert.Nil(t, err)

				// Check non-date fields
				assert.Equal(t, test.expectedBody.BookTitle, actualLoan.BookTitle, "Book title mismatch")
				assert.Equal(t, test.expectedBody.NameOfBorrower, actualLoan.NameOfBorrower, "Borrower name mismatch")

				// Check LoanDate is recent (within 5 seconds of current UTC time)
				nowUTC := time.Now().UTC()
				loanDateUTC := actualLoan.LoanDate.UTC()
				assert.WithinDuration(t, nowUTC, loanDateUTC, 5*time.Second, "LoanDate should be within 5 seconds of now")

				// Check ReturnDate is exactly 3 weeks after LoanDate
				expectedReturnDate := actualLoan.LoanDate.Add(time.Hour * 24 * 7 * 3)
				assert.WithinDuration(t, expectedReturnDate, actualLoan.ReturnDate, 5*time.Second, "ReturnDate should be 3 weeks after LoanDate")
			})
		}
	})

	t.Run("ReturnBook", func(t *testing.T) {
		tests := []struct {
			description  string
			route        string
			method       string
			requestBody  dto.BorrowBook
			expectedCode int
			expectedBody string
		}{
			{
				description:  "Successfully return the book",
				route:        "/Return",
				method:       http.MethodPost,
				requestBody:  dto.BorrowBook{Title: bookTitle},
				expectedCode: http.StatusOK,
				expectedBody: `{"status":"success"}`,
			},
		}

		for _, test := range tests {
			t.Run(test.description, func(t *testing.T) {
				reqBody, _ := json.Marshal(test.requestBody)
				req := httptest.NewRequest(test.method, test.route, bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				assert.Nil(t, err)
				assert.Equal(t, test.expectedCode, resp.StatusCode)

				body, err := io.ReadAll(resp.Body)
				assert.Nil(t, err)
				assert.JSONEq(t, test.expectedBody, string(body))
			})
		}
	})
}
