package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DarrelA/e-lib/config"
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/postgres"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	upperCaseBookTitle = "Anna"
	lowerCaseBookTitle = "anna"
	username           = "User1"
	userID             = 1
)

type mockSessionRepository struct{ mock.Mock }

func (m *mockSessionRepository) NewSession(userID int64) (string, *apperrors.RestErr) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return "", args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(string), nil
}

func (m *mockSessionRepository) GetSessionData(sessionID string) (*entity.Session, *apperrors.RestErr) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*entity.Session), nil
}

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) GetUserByID(userID int64) (dto.UserDetail, *apperrors.RestErr) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return dto.UserDetail{}, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(dto.UserDetail), nil
}

func (m *mockUserRepository) GetUserID(provider string, id string, email string) (int64, *apperrors.RestErr) {
	args := m.Called(provider, id, email)
	user_id, ok := args.Get(0).(int64)
	if !ok {
		return -1, args.Get(1).(*apperrors.RestErr)
	}
	return user_id, nil
}

func (m *mockUserRepository) SaveUser(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*entity.User), nil
}

type mockBookRepository struct {
	mock.Mock
	ExpectedBook *dto.BookDetail
}

func (m *mockBookRepository) GetBook(requestID string, title string) (*dto.BookDetail, *apperrors.RestErr) {
	args := m.Called(requestID, strings.ToLower(title))
	book, ok := args.Get(0).(*dto.BookDetail)
	if !ok {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return book, nil
}

type mockLoanRepository struct{ mock.Mock }

func (m *mockLoanRepository) BorrowBook(requestID string, userDetail dto.UserDetail, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	args := m.Called(requestID, userDetail, bookDetail.UUID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*dto.LoanDetail), nil
}

func (m *mockLoanRepository) ExtendBookLoan(requestID string, user_id int64, bookDetail *dto.BookDetail) (*dto.LoanDetail, *apperrors.RestErr) {
	args := m.Called(requestID, user_id, bookDetail)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*apperrors.RestErr)
	}
	return args.Get(0).(*dto.LoanDetail), nil
}

func (m *mockLoanRepository) ReturnBook(requestID string, user_id int64, book_uuid uuid.UUID) *apperrors.RestErr {
	args := m.Called(requestID, user_id, book_uuid)
	err := args.Get(0)
	if err == nil {
		return nil
	}
	return err.(*apperrors.RestErr)
}

func initializeEnv() *config.EnvConfig {
	envConfig := config.NewEnvConfig()
	envConfig.LoadServerConfig()
	envConfig.LoadOAuth2Config()
	envConfig.LoadPostgresConfig()
	config, _ := envConfig.(*config.EnvConfig)
	return config
}

func TestRoutes(t *testing.T) {
	// Shared setup
	testUser := entity.User{ID: userID, Name: username}
	bookUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	expectedBook := dto.BookDetail{UUID: bookUUID, Title: lowerCaseBookTitle, AvailableCopies: 10}

	now := time.Now().UTC()
	expectedLoan := dto.LoanDetail{
		BookTitle:      lowerCaseBookTitle,
		NameOfBorrower: username,
		LoanDate:       now,
		ReturnDate:     now.Add(time.Hour * 24 * 7 * 4),
	}

	extendedReturnDate := now.Add(time.Hour * 24 * 7 * 3)
	extendedLoan := dto.LoanDetail{
		BookTitle:      lowerCaseBookTitle,
		NameOfBorrower: username,
		LoanDate:       now,
		ReturnDate:     extendedReturnDate,
	}

	config := initializeEnv()

	config.PostgresDBConfig = &entity.PostgresDBConfig{
		Username:     "testuser",
		Password:     "testpassword",
		Host:         "localhost",
		Port:         "5432",
		Name:         "testdb",
		SslMode:      "disable",
		PoolMaxConns: "10",
	}
	postgresDBInstance := &postgres.PostgresDB{}

	mockUserRepo := new(mockUserRepository)
	mockSessionRepo := new(mockSessionRepository)
	googleOAuth2Service := interfaceSvc.NewGoogleOAuth2(config.OAuth2Config, mockUserRepo, mockSessionRepo)

	mockBookRepo := new(mockBookRepository)
	mockLoanRepo := new(mockLoanRepository)

	mockBookRepo.On("GetBook", mock.Anything, lowerCaseBookTitle).Return(&expectedBook, nil)
	mockLoanRepo.On("BorrowBook", mock.Anything, testUser, bookUUID).Return(&expectedLoan, nil)
	mockLoanRepo.On("ExtendBookLoan", mock.Anything, testUser.ID, &expectedBook).Return(&extendedLoan, nil)
	mockLoanRepo.On("ReturnBook", mock.Anything, testUser.ID, bookUUID).Return(nil, nil)

	bookService := interfaceSvc.NewBookService(mockBookRepo)
	loanService := interfaceSvc.NewLoanService(mockBookRepo, mockLoanRepo)

	mockNewSessionFunc := func(userID int64) (string, *apperrors.RestErr) {
		sessionID := "dummy_id"
		return sessionID, nil
	}

	mockSaveUserFunc := func(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr) {
		libUser := &entity.User{ID: userID, Name: username}
		return libUser, nil
	}

	mockGetSessionDataFunc := func(sessionID string) (*entity.Session, *apperrors.RestErr) {
		userIDString := strconv.FormatInt(userID, 10)       // Convert int64 to string
		createdAt := time.Now().Unix()                      // Get Unix timestamp
		createdAtString := strconv.FormatInt(createdAt, 10) // Convert Unix timestamp to string
		sessionData := &entity.Session{UserID: userIDString, CreatedAt: createdAtString}
		return sessionData, nil
	}

	mockGetUserByIDFunc := func(userID int64) (dto.UserDetail, *apperrors.RestErr) {
		userDetail := dto.UserDetail{ID: userID, Name: username}
		return userDetail, nil
	}

	app := NewRouter(
		config, googleOAuth2Service, postgresDBInstance,
		bookService, loanService,
		mockNewSessionFunc, mockSaveUserFunc,
		mockGetSessionDataFunc, mockGetUserByIDFunc,
	)

	t.Run("GetBookByTitle", func(t *testing.T) {
		tests := []struct {
			description  string
			route        string
			expectedCode int
			expectedBody string
		}{
			{
				description:  "Get existing book by title",
				route:        fmt.Sprintf("/Book?title=%s", upperCaseBookTitle),
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
			requestBody  dto.BookRequest
			expectedCode int
			expectedBody dto.LoanDetail
		}{
			{
				description:  "Successfully borrow book for 4 weeks",
				route:        "/Borrow",
				method:       http.MethodPost,
				requestBody:  dto.BookRequest{Title: upperCaseBookTitle},
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
			requestBody  dto.BookRequest
			expectedCode int
			expectedBody dto.LoanDetail
		}{
			{
				description:  "Successfully extend the loan of the book by 3 weeks",
				route:        "/Extend",
				method:       http.MethodPost,
				requestBody:  dto.BookRequest{Title: upperCaseBookTitle},
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
			requestBody  dto.BookRequest
			expectedCode int
			expectedBody string
		}{
			{
				description:  "Successfully return the book",
				route:        "/Return",
				method:       http.MethodPost,
				requestBody:  dto.BookRequest{Title: upperCaseBookTitle},
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
