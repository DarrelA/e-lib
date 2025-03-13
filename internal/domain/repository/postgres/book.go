package postgres

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
)

type BookRepository interface {
	GetBook(requestId string, title string) (*dto.BookDetail, *apperrors.RestErr)
}
