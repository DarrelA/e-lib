package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	errMsgEmptyTitle   = "User did not provide title."
	errMsgBookNotFound = "Book not found."
)

type BookService struct {
	Books []entity.Book
}

func NewBookService(books []entity.Book) appSvc.BookService {
	return &BookService{Books: books}
}

func (bs *BookService) GetBookByTitleHandler(c *fiber.Ctx) error {
	title := c.Query("title")
	if title == "" {
		log.Error().Msg(errMsgEmptyTitle)
		return c.Status(fiber.StatusBadRequest).JSON(errMsgEmptyTitle)
	}

	bookDetail, err := bs.GetBookByTitle(title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(bookDetail)
}

func (bs *BookService) GetBookByTitle(title string) (*dto.BookDetail, *apperrors.RestErr) {
	for _, book := range bs.Books {
		if book.Title == title {
			bookDetail := dto.BookDetail{
				UUID:            *book.UUID,
				Title:           book.Title,
				AvailableCopies: book.AvailableCopies,
			}
			return &bookDetail, nil
		}
	}

	return nil, apperrors.NewNotFoundError(errMsgBookNotFound)
}
