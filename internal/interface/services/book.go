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
	errMsgNotFoundOrIncorrectType = "%s not found or has incorrect type"
	errMsgBookNotFound            = "book not found"
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
		log.Error().Msgf(errMsgNotFoundOrIncorrectType, "bookTitleKey")
	}

	requestId, ok := c.Locals("requestid").(string)
	if !ok {
		log.Error().Msgf(errMsgNotFoundOrIncorrectType, "requestid")
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
