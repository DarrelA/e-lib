package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	pathToSqlFile = "./config/schema.elib.sql"

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

type SeedRepository struct{ dbpool *pgxpool.Pool }

func NewRepository(dbpool *pgxpool.Pool, user *entity.User) postgres.SeedRepository {
	ctx := context.Background()
	sqlData, err := os.ReadFile(pathToSqlFile)
	if err != nil {
		log.Error().Err(err).Msgf(errMsgUnableReadSchema, pathToSqlFile)
	}
	_, err = dbpool.Exec(ctx, string(sqlData))

	if err != nil {
		log.Error().Err(err).Msg(errMsgUnableToExecuteSQLScript)
	}

	_, err = dbpool.Exec(ctx, insertDummyUser, user.Name, user.Email)
	if err != nil {
		log.Error().Err(err).Msg("Error inserting dummy user")
		dbpool.Close()
		return nil
	}

	log.Info().Msg("successfully created the tables")
	return &SeedRepository{dbpool}
}

func (sr SeedRepository) SeedBooks(pathToBooksJsonFile string) error {
	content, err := os.ReadFile(pathToBooksJsonFile)
	if err != nil {
		log.Error().Err(err).Msg("")
		return err
	}

	var books []*entity.Book
	json.Unmarshal(content, &books)

	ctx := context.Background()
	tx, err := sr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgTransactionError)
		return fmt.Errorf(errMsgTransactionError, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	for _, book := range books {
		_, err = tx.Exec(ctx, insertBooksStmt, strings.ToLower(book.Title), book.AvailableCopies)
		if err != nil {
			log.Error().Err(err).Msg(errMsgInsertBooks)
			return fmt.Errorf(errMsgInsertBooks, book.Title, err)
		}
	}

	log.Info().Msg("Books seeded successfully.")
	return nil
}
