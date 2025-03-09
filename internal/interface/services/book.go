package services

import (
	"strings"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	repository "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	errMsgEmptyTitle   = "User did not provide title."
	errMsgBookNotFound = "Book not found."
)

type BookService struct {
	bookPGDB repository.BookRepository
}

func NewBookService(bookPGDB repository.BookRepository) appSvc.BookService {
	return &BookService{bookPGDB}
}

func (bs *BookService) GetBookByTitleHandler(c *fiber.Ctx) error {
	title := c.Query("title")
	if title == "" {
		log.Error().Msg(errMsgEmptyTitle)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": errMsgEmptyTitle})
	}

	bookDetail, err := bs.GetBookByTitle(title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(bookDetail)
}

func (bs *BookService) GetBookByTitle(title string) (*dto.BookDetail, *apperrors.RestErr) {
	bookDetail, err := bs.bookPGDB.GetBook(strings.ToLower(title))
	if err != nil {
		return nil, apperrors.NewNotFoundError(errMsgBookNotFound)
	}

	return bookDetail, nil
}
