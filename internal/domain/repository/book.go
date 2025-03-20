package repository

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
)

type BookRepository interface {
	GetBook(requestID string, title string) (*dto.BookDetail, *apperrors.RestErr)
}
