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
