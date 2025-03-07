package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/filedb"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/DarrelA/e-lib/internal/interface/transport/rest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	books := filedb.LoadBooksJsonData()
	user := getDummyUserData()

	bookService := interfaceSvc.NewBookService(books)
	loanService := interfaceSvc.NewLoanService(user, bookService)
	appInstance := rest.NewRouter(bookService, loanService)
	rest.StartServer(appInstance, "3000")
}

func getDummyUserData() entity.User {
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
	userDetail := entity.User{
		ID:        1,
		Name:      myInfoResponse.Name.Value,
		Email:     myInfoResponse.Email.Value,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	return userDetail
}
