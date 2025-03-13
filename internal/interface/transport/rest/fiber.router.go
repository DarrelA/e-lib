package rest

import (
	"github.com/DarrelA/e-lib/config"
	"github.com/DarrelA/e-lib/internal/apperrors"
	appSvc "github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/postgres"
	mw "github.com/DarrelA/e-lib/internal/interface/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	config *config.EnvConfig,
	postgresDBInstance *postgres.PostgresDB,
	bookService appSvc.BookService,
	loanService appSvc.LoanService,
) *fiber.App {
	log.Info().Msg("creating fiber instances")
	appInstance := fiber.New()

	log.Info().Msg("connecting middlewares")
	useMiddlewares(appInstance, config.AppEnv, postgresDBInstance.Dbpool)

	log.Info().Msg("setting up routes")
	appInstance.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	})

	/********************
	 *   BookServices   *
	 ********************/
	appInstance.Get("/Book", bookService.GetBookByTitleHandler)

	/********************
	*   LoanServices   *
	********************/
	appInstance.Post("/Borrow", loanService.BorrowBookHandler)
	appInstance.Post("/Extend", loanService.ExtendBookLoanHandler)
	appInstance.Post("/Return", loanService.ReturnBookHandler)

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

func useMiddlewares(appInstance *fiber.App, appEnv string, dbpool *pgxpool.Pool) {
	appInstance.Use(func(c *fiber.Ctx) error {
		c.Locals("appEnv", appEnv)
		c.Locals("dbpool", dbpool)
		return c.Next()
	})

	appInstance.Use(requestid.New())
	appInstance.Use(mw.Logger)
	appInstance.Use(mw.InputValidator)
	if appEnv == "test" {
		appInstance.Use(mw.SaveTestResToDB)
	}
}
