package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	repository "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/gofiber/fiber/v2"
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
	bookRequest := c.Locals("bookTitleKey").(dto.BookRequest)
	bookDetail, err := bs.GetBookByTitle(bookRequest)
	if err != nil {
		return c.Status(err.Status).JSON(err)
	}

	return c.Status(fiber.StatusOK).JSON(bookDetail)
}

func (bs *BookService) GetBookByTitle(bookRequest dto.BookRequest) (*dto.BookDetail, *apperrors.RestErr) {
	bookDetail, err := bs.bookPGDB.GetBook(bookRequest.Title)
	if err != nil {
		return nil, apperrors.NewNotFoundError(errMsgBookNotFound)
	}

	return bookDetail, nil
}
