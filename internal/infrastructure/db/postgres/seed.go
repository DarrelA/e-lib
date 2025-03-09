package postgres

import (
	"context"
	"os"

	"github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	pathToSqlFile                  = "./config/schema.elib.sql"
	errMsgUnableReadSchema         = "unable to read %s"
	errMsgUnableToExecuteSQLScript = "unable to execute sql script"
	errMsgUnableToLoadJSONFile     = "unable to load [%s]"
)

type SeedRepository struct{ dbpool *pgxpool.Pool }

func NewRepository(dbpool *pgxpool.Pool) postgres.SeedRepository {
	ctx := context.Background()
	sqlData, err := os.ReadFile(pathToSqlFile)
	if err != nil {
		log.Error().Err(err).Msgf(errMsgUnableReadSchema, pathToSqlFile)
	}
	_, err = dbpool.Exec(ctx, string(sqlData))

	if err != nil {
		log.Error().Err(err).Msg(errMsgUnableToExecuteSQLScript)
	}

	log.Info().Msg("successfully created the tables")
	return &SeedRepository{dbpool}
}
