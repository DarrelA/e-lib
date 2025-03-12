package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func SaveTestResToDB(c *fiber.Ctx) error {
	dbpool, ok := c.Locals("dbpool").(*pgxpool.Pool)
	if !ok {
		err := apperrors.NewInternalServerError("dbpool not found in fiber context")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	statusCode := c.Response().StatusCode()
	reqUrlQueryString, ok := c.Locals("bookTitleKey").(dto.BookRequest)
	if !ok {
		err := apperrors.NewInternalServerError("bookTitleKey not found or has incorrect type.")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	reqBody := string(c.Body())
	// Call c.Next() to execute the next middleware/handler *before* capturing the response.
	if err := c.Next(); err != nil {
		log.Error().Err(err).Msg("")
		return c.Status(fiber.StatusInternalServerError).JSON(err)
	}

	resBody := string(c.Response().Body()) // Capture response body *after* c.Next()

	// Start a database transaction
	ctx := c.Context()
	tx, err := dbpool.Begin(ctx)
	if err != nil {
		err := apperrors.NewInternalServerError("error starting transaction")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			log.Info().Msg("Transaction rolled back due to error")
		} else {
			tx.Commit(ctx)
			log.Info().Msg("Transaction committed successfully")
		}
	}()

	// Get the next expectedId by counting the rows in the Expected table
	expectedId, err := getNextExpectedId(ctx, tx)

	actual := entity.Actual{
		ExpectedId:        expectedId,
		StatusCode:        statusCode,
		ReqUrlQueryString: reqUrlQueryString.Title,
		ReqBody:           reqBody,
		ResBody:           resBody,
		CreatedAt:         time.Now(),
	}

	if err := insertActual(ctx, tx, actual); err != nil {
		err := apperrors.NewInternalServerError("error inserting into actual table")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	return nil
}

func insertActual(ctx context.Context, tx pgx.Tx, actual entity.Actual) error {
	query := `
		INSERT INTO Actual (expected_id, status_code, req_url_query_string, req_body, res_body, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := tx.Exec(ctx, query, actual.ExpectedId, actual.StatusCode, actual.ReqUrlQueryString, actual.ReqBody, actual.ResBody, actual.CreatedAt)
	return err
}

func getNextExpectedId(ctx context.Context, tx pgx.Tx) (int, error) {
	query := `SELECT COUNT(*) FROM Expected`
	var count int
	log.Info().Msgf("expectedId: %d", count)
	err := tx.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get next expected_id: %w", err)
	}
	return count + 1, nil // Increment to get the next ID
}
