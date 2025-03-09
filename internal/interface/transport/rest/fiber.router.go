package rest

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	mw "github.com/DarrelA/e-lib/internal/interface/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	errMsgStartServerFailure = "failed to start server"
)

func StartServer(app *fiber.App, port string) {
	log.Info().Msg("listening at port: " + port)
	err := app.Listen(":" + port)
	if err != nil {
		log.Error().Err(err).Msg(errMsgStartServerFailure)
	}
}

func NewRouter(
	bookService appSvc.BookService,
	loanService appSvc.LoanService,
) *fiber.App {
	log.Info().Msg("creating fiber instances")
	appInstance := fiber.New()

	log.Info().Msg("connecting middlewares")
	appInstance.Use(mw.Logger)

	log.Info().Msg("setting up routes")
	appInstance.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	})

	/********************
	 *   BookServices   *
	 ********************/
	appInstance.Get("/Book", mw.InputValidatorForGET, bookService.GetBookByTitleHandler)

	/********************
	*   LoanServices   *
	********************/
	appInstance.Post("/Borrow", mw.InputValidatorForPOST, loanService.BorrowBookHandler)
	appInstance.Post("/Extend", mw.InputValidatorForPOST, loanService.ExtendBookLoanHandler)
	appInstance.Post("/Return", mw.InputValidatorForPOST, loanService.ReturnBookHandler)

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
	return appInstance
}
