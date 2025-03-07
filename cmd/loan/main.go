package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/DarrelA/e-lib/internal/interface/transport/rest"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	booksJsonFilePath = "./testdata/json/books.json"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	books := loadBooksJsonData()
	user := getDummyUserData()

	bookService := interfaceSvc.NewBookService(books)
	loanService := interfaceSvc.NewLoanService(user, bookService)
	appInstance := rest.NewRouter(bookService, loanService)
	rest.StartServer(appInstance, "3000")
}

func loadBooksJsonData() []entity.BookDetail {
	jsonData, err := os.ReadFile(booksJsonFilePath)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	var books []entity.BookDetail
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

func getDummyUserData() entity.UserDetail {
	url := "https://sandbox.api.myinfo.gov.sg/com/v4/person-sample/S9812381D"
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	var myInfoResponse entity.MyInfoResponse
	err = json.Unmarshal(body, &myInfoResponse)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	currentTime := time.Now()
	userDetail := entity.UserDetail{
		ID:        1,
		Name:      myInfoResponse.Name.Value,
		Email:     myInfoResponse.Email.Value,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	return userDetail
}
