// postgres.go
package postgres

import (
	"context"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type BookRepository struct {
	dbpool *pgxpool.Pool
}

func NewBookRepository(dbpool *pgxpool.Pool) repository.BookRepository {
	return &BookRepository{dbpool}
}

func (br BookRepository) GetBook(requestId string, title string) (*dto.BookDetail, *apperrors.RestErr) {
	bookDetail := &dto.BookDetail{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, entity.RequestIdKey, requestId)

	const queryGetBook = "SELECT uuid, title, available_copies FROM books WHERE title=$1;"
	err := br.dbpool.QueryRow(ctx, queryGetBook, title).
		Scan(&bookDetail.UUID, &bookDetail.Title, &bookDetail.AvailableCopies)

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Ctx(ctx).Error().Err(err).Msg(errMsgContextTimeout)
			return nil, apperrors.NewInternalServerError(errMsgContextTimeout)
		}

		if err == pgx.ErrNoRows {
			return nil, apperrors.NewBadRequestError(err.Error())
		}

		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(err.Error())
	}

	return bookDetail, nil
}
