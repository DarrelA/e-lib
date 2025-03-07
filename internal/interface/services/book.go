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
	Books []entity.BookDetail
}

func NewBookService(books []entity.BookDetail) appSvc.BookService {
	return &BookService{Books: books}
}

func (bs *BookService) GetBookByTitleHandler(c *fiber.Ctx) error {
	title := c.Query("title")
	if title == "" {
		log.Error().Msg(errMsgEmptyTitle)
		return c.Status(fiber.StatusBadRequest).JSON(errMsgEmptyTitle)
	}

	book, err := bs.GetBookByTitle(title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	response := dto.BookTitleAvailability{
		Title:           book.Title,
		AvailableCopies: book.AvailableCopies,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (bs *BookService) GetBookByTitle(title string) (*entity.BookDetail, *apperrors.RestErr) {
	for _, book := range bs.Books {
		if book.Title == title {
			return &book, nil
		}
	}

	return nil, apperrors.NewNotFoundError(errMsgBookNotFound)
}
