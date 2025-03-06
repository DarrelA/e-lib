package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type BookService struct {
	Books []entity.BookDetail
}

func NewBookService(books []entity.BookDetail) appSvc.BookService {
	return &BookService{Books: books}
}

func (bs *BookService) GetBookByTitleHandler(c *fiber.Ctx) error {
	title := c.Params("title")
	if title == "" {
		err := apperrors.NewBadRequestError(apperrors.ErrMsgSomethingWentWrong)
		log.Error().Err(err).Msg(apperrors.ErrMsgSomethingWentWrong)
	}

	book, err := bs.GetBookByTitle(title)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(book)
}

func (bs *BookService) GetBookByTitle(title string) (*entity.BookDetail, *apperrors.RestErr) {
	for _, book := range bs.Books {
		if book.Title == title {
			return &book, nil
		}
	}

	return nil, apperrors.NewBadRequestError(apperrors.ErrMsgBookNotFound)
}
