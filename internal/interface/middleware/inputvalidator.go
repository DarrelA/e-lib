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
		errorMessages = append(errorMessages, fmt.Sprintf("Field %s: %s", e.Field(), e.Tag()))
	}

	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Validation failed: %v", errorMessages)})
}

func InputValidatorForGET(c *fiber.Ctx) error {
	bookTitle := new(dto.BookRequest)
	if err := c.QueryParser(bookTitle); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameters"})
	}

	if err := validateBookTitle(*bookTitle); err != nil {
		return handleValidationError(c, err)
	}

	bookTitle.Title = strings.ToLower(bookTitle.Title)
	c.Locals("bookTitleKey", *bookTitle) // Use dereferenced value
	return c.Next()
}

func InputValidatorForPOST(c *fiber.Ctx) error {
	var bookTitle dto.BookRequest
	if err := c.BodyParser(&bookTitle); err != nil {
		log.Error().Err(err).Msg("")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
	}

	if err := validateBookTitle(bookTitle); err != nil {
		return handleValidationError(c, err)
	}

	bookTitle.Title = strings.ToLower(bookTitle.Title)
	c.Locals("bookTitleKey", bookTitle)
	return c.Next()
}
