package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/application/repository"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockJsonFileRepo struct {
	repository.JsonFileRepository
}

func (m *mockJsonFileRepo) LoadLoanDetails() ([]*entity.Loan, error) {
	return []*entity.Loan{}, nil // Simulate empty loans
}

func (m *mockJsonFileRepo) SaveLoanDetail(loan *entity.Loan) error {
	return nil // Simulate successful save
}

func (m *mockJsonFileRepo) DecrementAvailableCopies(title string) error {
	return nil // Simulate decrement
}

func TestRoutes(t *testing.T) {
	// Shared setup
	testUser := entity.User{ID: 1, Name: "User1"}
	bookUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testBook := entity.Book{UUID: &bookUUID, Title: "Anna", AvailableCopies: 10}
	bookService := interfaceSvc.NewBookService([]entity.Book{testBook})
	mockRepo := &mockJsonFileRepo{}
	loanService := interfaceSvc.NewLoanService(testUser, bookService, mockRepo)
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
				route:        "/Book?title=Anna",
				expectedCode: http.StatusOK,
				expectedBody: `{"uuid":"123e4567-e89b-12d3-a456-426614174000","title":"Anna","available_copies":10}`,
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
		expectedLoan := dto.LoanDetail{
			BookTitle:      "Anna",
			NameOfBorrower: "User1",
			LoanDate:       time.Time{},
			ReturnDate:     time.Time{},
		}

		tests := []struct {
			description  string
			route        string
			method       string
			requestBody  dto.BorrowBook
			expectedCode int
			expectedBody dto.LoanDetail
		}{
			{
				description:  "Successfully borrow book",
				route:        "/Borrow",
				method:       http.MethodPost,
				requestBody:  dto.BorrowBook{Title: "Anna"},
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
}
