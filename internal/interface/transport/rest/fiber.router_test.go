package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Fixed timestamps
var (
	loanDate   = time.Date(2024, 3, 8, 11, 0, 0, 0, time.FixedZone("UTC+8", 8*60*60))
	returnDate = loanDate.Add(time.Hour * 24 * 7 * 4) // 4 weeks
)

type mockLoanService struct {
	user        entity.User
	bookService appSvc.BookService
}

func (m *mockLoanService) BorrowBookHandler(c *fiber.Ctx) error {
	var borrowBook dto.BorrowBook
	if err := c.BodyParser(&borrowBook); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON("Invalid request Body")
	}

	loanDetail, err := m.BorrowBook(borrowBook.Title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}
	return c.Status(fiber.StatusOK).JSON(loanDetail)
}

func (m *mockLoanService) BorrowBook(title string) (*dto.LoanDetail, *apperrors.RestErr) {
	return &dto.LoanDetail{
		BookTitle:      title,
		NameOfBorrower: m.user.Name,
		LoanDate:       loanDate,
		ReturnDate:     returnDate,
	}, nil
}

func (m *mockLoanService) ExtendBookLoanHandler(c *fiber.Ctx) error { return nil }
func (m *mockLoanService) ExtendBookLoan(title string) (*dto.LoanDetail, *apperrors.RestErr) {
	return nil, nil
}

func (m *mockLoanService) ReturnBookHandler(c *fiber.Ctx) error       { return nil }
func (m *mockLoanService) ReturnBook(title string) *apperrors.RestErr { return nil }

func TestGetBookByTitle(t *testing.T) {
	// Setup test data with dummy UUID
	bookUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testBook := entity.Book{
		UUID:            &bookUUID,
		Title:           "Anna",
		AvailableCopies: 10,
	}

	bookService := interfaceSvc.NewBookService([]entity.Book{testBook})
	loanService := &mockLoanService{}
	app := NewRouter(bookService, loanService)

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
		// Create httptest request
		req := httptest.NewRequest("GET", test.route, nil)

		// Use Fiber's Test method with httptest recorder
		resp, err := app.Test(req)
		assert.Nilf(t, err, test.description)
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)

		body, err := io.ReadAll(resp.Body)
		assert.Nilf(t, err, test.description)
		assert.JSONEqf(t, test.expectedBody, string(body), test.description)
	}
}

func TestBorrowBook(t *testing.T) {
	testUser := entity.User{
		ID:   1,
		Name: "User1",
	}

	bookUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	testBook := entity.Book{
		UUID:            &bookUUID,
		Title:           "Anna",
		AvailableCopies: 10,
	}

	bookService := interfaceSvc.NewBookService([]entity.Book{testBook})
	loanService := &mockLoanService{
		user:        testUser,
		bookService: bookService,
	}

	app := NewRouter(bookService, loanService)

	expectedLoan := dto.LoanDetail{
		BookTitle:      "Anna",
		NameOfBorrower: "User1",
		LoanDate:       loanDate,
		ReturnDate:     returnDate,
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
		reqBody, _ := json.Marshal(test.requestBody)
		req := httptest.NewRequest(test.method, test.route, bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.Nilf(t, err, test.description)
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var actualLoan dto.LoanDetail
		err = json.Unmarshal(body, &actualLoan)
		assert.Nilf(t, err, test.description)

		// Compare individual fields with time equality check
		assert.Equal(t, test.expectedBody.BookTitle, actualLoan.BookTitle, "Book title mismatch")
		assert.Equal(t, test.expectedBody.NameOfBorrower, actualLoan.NameOfBorrower, "Borrower name mismatch")
		assert.True(t, test.expectedBody.LoanDate.Equal(actualLoan.LoanDate),
			"LoanDate mismatch (expected %s, got %s)", test.expectedBody.LoanDate, actualLoan.LoanDate)
		assert.True(t, test.expectedBody.ReturnDate.Equal(actualLoan.ReturnDate),
			"ReturnDate mismatch (expected %s, got %s)", test.expectedBody.ReturnDate, actualLoan.ReturnDate)
	}
}
