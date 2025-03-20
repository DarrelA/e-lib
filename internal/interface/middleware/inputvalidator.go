package middleware

import (
	"fmt"
	"strings"

	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func validateBookTitle(bookTitle dto.BookRequest) error {
	return validate.Struct(bookTitle)
}

func handleValidationError(c *fiber.Ctx, err error) error {
	validationErrors := err.(validator.ValidationErrors)
	errorMessages := make([]string, 0, len(validationErrors))

	for _, e := range validationErrors {
		errorMessages = append(errorMessages, fmt.Sprintf("field %s: %s", e.Field(), e.Tag()))
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("validation failed: %v", errorMessages)})
}

func InputValidator(c *fiber.Ctx) error {
	var bookTitle dto.BookRequest
	var err error

	switch c.Method() {
	case "GET":
		bookTitlePtr := new(dto.BookRequest) // Use pointer for QueryParser to correctly bind
		err = c.QueryParser(bookTitlePtr)    // Parse into pointer
		if err == nil {
			bookTitle = *bookTitlePtr // Dereference the pointer
		}

	case "POST":
		err = c.BodyParser(&bookTitle)

	default:
		return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{"message": "method not allowed"})
	}

	if err != nil {
		log.Error().Err(err).Msg("error parsing request body/query")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "invalid request parameters"})
	}

	if err := validateBookTitle(bookTitle); err != nil {
		return handleValidationError(c, err)
	}

	bookTitle.Title = strings.ToLower(bookTitle.Title)
	c.Locals("bookTitleKey", bookTitle)
	return c.Next()
}
