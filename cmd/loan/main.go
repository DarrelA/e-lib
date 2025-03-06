package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	jsonData, err := os.ReadFile("./testdata/json/books.json")
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

	for _, book := range books {
		fmt.Printf("Title: %s, Copies: %d, UUID: %s, CreatedAt: %s\n", book.Title, book.AvailableCopies, book.UUID, book.CreatedAt)
	}
}
