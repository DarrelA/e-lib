package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DarrelA/e-lib/config"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type SeedRepository struct {
	config *config.EnvConfig
	dbpool *pgxpool.Pool
}

func NewRepository(config *config.EnvConfig, dbpool *pgxpool.Pool) repository.SeedRepository {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var schemasToExecute []string
	schemasToExecute = append(schemasToExecute, config.PathToSQLSchema)

	for _, schemaPath := range schemasToExecute {
		if err := executeSchema(ctx, dbpool, schemaPath); err != nil {
			dbpool.Close()
			return nil
		}
	}

	log.Info().Msgf("successfully created the tables using %s", config.PathToSQLSchema)
	return &SeedRepository{config, dbpool}
}

func executeSchema(ctx context.Context, dbpool *pgxpool.Pool, schemaPath string) error {
	sqlData, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Error().Err(err).Msgf("unable to read %s", schemaPath)
		return err
	}
	_, err = dbpool.Exec(ctx, string(sqlData))
	if err != nil {
		log.Error().Err(err).Msg("unable to execute sql script")
		return err
	}
	return nil
}

func (sr SeedRepository) SeedBooks() error {
	content, err := os.ReadFile(sr.config.PathToBooksJsonFile)
	if err != nil {
		log.Error().Err(err).Msg("failed to read books JSON file")
		return err
	}

	var books []*entity.Book
	err = json.Unmarshal(content, &books)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal books JSON")
		return err
	}

	err = sr.executeTransaction(func(tx pgx.Tx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		const insertBooksStmt = `
			INSERT INTO Books (uuid, title, available_copies, created_at, updated_at)
			VALUES (gen_random_uuid(), lower($1), $2, NOW(), NOW())
			ON CONFLICT (title) DO NOTHING;  -- Skip duplicates based on title
		`
		for _, book := range books {
			_, err := tx.Exec(ctx, insertBooksStmt, strings.ToLower(book.Title), book.AvailableCopies)
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("error inserting book '%s'", book.Title))
				return fmt.Errorf("error inserting book '%s': %w", book.Title, err)
			}
		}

		log.Info().Msgf("books seeded successfully using %s", sr.config.PathToBooksJsonFile)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

type TxFunc func(pgx.Tx) error

func (sr SeedRepository) executeTransaction(txFunc TxFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := sr.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToBeginTransaction)
		return fmt.Errorf(errMsgFailedToBeginTransaction+": %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			log.Error().Interface("panic", p).Msg("panic occurred during transaction") // Log the panic
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
		} else if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
			}
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				log.Error().Err(err).Msg(errMsgFailedToCommitTransaction)
			}
		}
	}()

	err = txFunc(tx)
	return err
}
