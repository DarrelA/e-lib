package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
)

// BookService defines the interface for managing book-related operations (e.g., retrieving book information).
type BookService interface {
	GetBookByTitle(title string) (*entity.BookDetail, *apperrors.RestErr)
}
