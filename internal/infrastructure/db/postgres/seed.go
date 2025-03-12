package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DarrelA/e-lib/config"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	pathToSQLTestSchema              = "./config/test.schema.api_testing.sql"
	pathToBooksExpectedRes           = "/root/testdata/json/test.booksExpectedRes.json"
	pathToCompareTestReqAndResReport = "/root/testdata/reports/it_report.json"

	errMsgUnableReadSchema         = "unable to read %s"
	errMsgUnableToExecuteSQLScript = "unable to execute sql script"
	errMsgUnableToLoadJSONFile     = "unable to load [%s]"

	errMsgTransactionError = "error starting transaction: %w"
	errMsgInsertBooks      = "error inserting book '%s': %w"
)

const (
	insertDummyUser = `
		INSERT INTO users (name, email, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING;  -- Skip duplicates based on email
	`

	insertBooksStmt = `
		INSERT INTO Books (uuid, title, available_copies, created_at, updated_at)
		VALUES (gen_random_uuid(), lower($1), $2, NOW(), NOW())
		ON CONFLICT (title) DO NOTHING;  -- Skip duplicates based on title
	`

	insertBooksExpectedResStmt = `
		INSERT INTO Expected (method, url_path, status_code, res_body_contains)
		VALUES ($1, $2, $3, $4)
	`
)

type SeedRepository struct {
	config *config.EnvConfig
	dbpool *pgxpool.Pool
}

func NewRepository(config *config.EnvConfig, dbpool *pgxpool.Pool, user *entity.User) postgres.SeedRepository {
	ctx := context.Background()
	var schemasToExecute []string
	schemasToExecute = append(schemasToExecute, config.PathToSQLSchema)
	if config.AppEnv == "test" {
		schemasToExecute = append(schemasToExecute, pathToSQLTestSchema)
	}

	for _, schemaPath := range schemasToExecute {
		if err := executeSchema(ctx, dbpool, schemaPath); err != nil {
			dbpool.Close()
			return nil
		}
	}

	log.Info().Msgf("successfully created the tables using %s", config.PathToSQLSchema)

	_, err := dbpool.Exec(ctx, insertDummyUser, user.Name, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Error inserting dummy user")
		dbpool.Close()
		return nil
	}

	log.Info().Msgf("successfully inserted dummy user %s", user.Name)
	return &SeedRepository{config, dbpool}
}

func executeSchema(ctx context.Context, dbpool *pgxpool.Pool, schemaPath string) error {
	sqlData, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Error().Err(err).Msgf(errMsgUnableReadSchema, schemaPath)
		return err
	}
	_, err = dbpool.Exec(ctx, string(sqlData))
	if err != nil {
		log.Error().Err(err).Msg(errMsgUnableToExecuteSQLScript)
		return err
	}
	return nil
}

func (sr SeedRepository) SeedBooks() error {
	content, err := os.ReadFile(sr.config.PathToBooksJsonFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read books JSON file")
		return err
	}

	var books []*entity.Book
	err = json.Unmarshal(content, &books)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal books JSON")
		return err
	}

	err = sr.executeTransaction(func(tx pgx.Tx) error {
		ctx := context.Background()

		for _, book := range books {
			_, err := tx.Exec(ctx, insertBooksStmt, strings.ToLower(book.Title), book.AvailableCopies)
			if err != nil {
				log.Error().Err(err).Msg(errMsgInsertBooks)
				return fmt.Errorf(errMsgInsertBooks, book.Title, err)
			}
		}

		log.Info().Msgf("Books seeded successfully using %s.", sr.config.PathToBooksJsonFile)

		if sr.config.AppEnv == "test" {
			content, err := os.ReadFile(pathToBooksExpectedRes)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read expected results JSON file")
				return err
			}

			var expected []*entity.Expected
			err = json.Unmarshal(content, &expected)
			if err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal expected results JSON")
				return err
			}

			if len(expected) == 0 {
				log.Warn().Msg("No expected results found in JSON, skipping expected results seeding.")
				return nil // or return error, depending on your needs
			}

			for _, exp := range expected {
				_, err = tx.Exec(ctx, insertBooksExpectedResStmt, exp.Method, exp.UrlPath, exp.StatusCode, exp.ResBodyContains)
				if err != nil {
					log.Error().Err(err).Msg(errMsgUnableToExecuteSQLScript)
					return err
				}
			}

			log.Info().Msgf("ExpectedResults seeded successfully using %s.", pathToBooksExpectedRes)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type TxFunc func(pgx.Tx) error

func (sr SeedRepository) executeTransaction(txFunc TxFunc) error {
	ctx := context.Background()
	tx, err := sr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgTransactionError)
		return fmt.Errorf(errMsgTransactionError, err)
	}

	defer func() {
		if p := recover(); p != nil {
			log.Error().Interface("panic", p).Msg("Panic occurred during transaction") // Log the panic
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg("Failed to rollback transaction after panic")
			}
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg("Failed to rollback transaction")
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to commit transaction")
			}
		}
	}()

	err = txFunc(tx)
	return err
}

func (sr SeedRepository) CompareTestReqAndRes() {
	ctx := context.Background()

	query := "SELECT e.id, e.method, e.url_path, a.req_url_query_string, a.req_body, e.status_code, a.status_code, a.res_body, e.res_body_contains, a.created_at FROM expected e JOIN actual a ON e.id = a.expected_id"
	rows, err := sr.dbpool.Query(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query expected and actual data")
		return
	}
	defer rows.Close()

	var results []entity.CompareResult

	for rows.Next() {
		var result entity.CompareResult
		var expected entity.Expected
		var actual entity.Actual

		err := rows.Scan(
			&expected.Id,
			&expected.Method,
			&expected.UrlPath,
			&actual.ReqUrlQueryString,
			&actual.ReqBody,
			&expected.StatusCode,
			&actual.StatusCode,
			&actual.ResBody,
			&expected.ResBodyContains,
			&actual.CreatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan row")
			continue
		}

		result.Id = expected.Id
		result.Method = expected.Method
		result.UrlPath = expected.UrlPath
		result.ReqUrlQueryString = actual.ReqUrlQueryString
		result.ReqBody = actual.ReqBody
		result.ExpectedStatusCode = expected.StatusCode
		result.ActualStatusCode = actual.StatusCode
		result.ResBody = actual.ResBody
		result.ResBodyContains = expected.ResBodyContains
		result.CreatedAt = actual.CreatedAt
		result.Reason = []string{}

		statusCodePass, statusCodeReason := assertEqual(fmt.Sprintf("%d", expected.StatusCode),
			fmt.Sprintf("%d", actual.StatusCode), "Status code mismatch")
		resBodyContainsPass, resBodyContainsReason := true, ""

		if expected.ResBodyContains != "" {
			resBodyContainsPass, resBodyContainsReason = assertContains(actual.ResBody, expected.ResBodyContains,
				"Response body does not contain expected string")
		}

		if !statusCodePass {
			result.Reason = append(result.Reason, statusCodeReason)
		}
		if !resBodyContainsPass && resBodyContainsReason != "" {
			result.Reason = append(result.Reason, resBodyContainsReason)
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating over rows")
		return
	}

	file, err := os.Create(pathToCompareTestReqAndResReport)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to create file: %s", pathToCompareTestReqAndResReport)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Add indentation for readability

	if err := encoder.Encode(results); err != nil {
		log.Error().Err(err).Msg("Failed to encode results to JSON")
		return
	}

	log.Info().Msgf("Comparison results written to %s", pathToCompareTestReqAndResReport)
}

func assertEqual(expected string, actual string, msg string) (bool, string) {
	if expected == actual {
		return true, ""
	}

	reason := fmt.Sprintf("assertion failed: Expected [%s], Actual [%s]. %s", expected, actual, msg)
	return false, reason
}

func assertContains(actual string, expected string, msg string) (bool, string) {
	if strings.Contains(actual, expected) {
		return true, ""
	}

	reason := fmt.Sprintf("assertion failed:  Expected string [%s] to contain substring [%s]. %s", actual, expected, msg)
	return false, reason
}
