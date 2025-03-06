package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
)

// BookService defines the interface for managing book-related operations (e.g., retrieving book information).
type BookService interface {
	GetBookByTitleHandler(c *fiber.Ctx) error
	GetBookByTitle(title string) (*entity.BookDetail, *apperrors.RestErr)
}
