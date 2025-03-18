package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/repository"
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
	bookRequest, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		log.Error().Msg("bookTitleKey not found or has incorrect type.")
	}

	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msg("requestid not found or has incorrect type.")
	}

	bookDetail, err := bs.GetBookByTitle(requestId, bookRequest)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(bookDetail)
}

func (bs *BookService) GetBookByTitle(requestId string, bookRequest dto.BookRequest) (*dto.BookDetail, *apperrors.RestErr) {
	bookDetail, err := bs.bookPGDB.GetBook(requestId, bookRequest.Title)
	if err != nil {
		return nil, apperrors.NewNotFoundError(errMsgBookNotFound)
	}

	return bookDetail, nil
}
