package rest

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func StartServer(app *fiber.App, port string) {
	log.Info().Msg("listening at port: " + port)
	err := app.Listen(":" + port)
	if err != nil {
		log.Error().Err(err).Msg(apperrors.ErrMsgStartServerFailure)
	}
}

func NewRouter(
	bookService appSvc.BookService,
) *fiber.App {
	log.Info().Msg("creating fiber instances")
	appInstance := fiber.New()
	libInstance := fiber.New()
	appInstance.Mount("/lib", libInstance)

	log.Info().Msg("setting up routes")
	v1 := libInstance.Group("/api/v1", func(c *fiber.Ctx) error { // middleware for /api/v1
		c.Set("Version", "v1")
		return c.Next()
	})

	v1.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	})

	/********************
	 *   BookServices   *
	 ********************/
	book := v1.Group("/books")
	book.Get("/:title", bookService.GetBookByTitleHandler)

	appInstance.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		err := apperrors.NewBadRequestError("Invalid Path: " + path)
		log.Error().Err(err).Msg("")
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": "404 - Not Found",
		})
	})

	log.Info().Msg("/health endpoint is available")
	log.Debug().Msgf("appInstance memory address: %p", appInstance)
	log.Debug().Msgf("authServiceInstance memory address: %p", libInstance)
	return appInstance
}
