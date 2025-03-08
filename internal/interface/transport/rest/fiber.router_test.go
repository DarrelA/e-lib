package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockLoanService struct{}

func (m *mockLoanService) BorrowBookHandler(c *fiber.Ctx) error { return nil }
func (m *mockLoanService) BorrowBook(title string) (*dto.LoanDetail, *apperrors.RestErr) {
	return nil, nil
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
