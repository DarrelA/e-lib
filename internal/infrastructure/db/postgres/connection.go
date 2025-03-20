package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	errMsgContextTimeout              = "context timeout occurred"
	errMsgFailedToBeginTransaction    = "failed to begin transaction"
	errMsgFailedToRollbackTransaction = "failed to rollback transaction"
	errMsgFailedToCommitTransaction   = "failed to commit transaction"

	infoMsgRollbackTransactionSuccess  = "transaction rollback successfully"
	infoMsgCommittedTransactionSuccess = "transaction committed successfully"
)

/*
dbpool is the database connection pool.
Package pgxpool is a concurrency-safe connection pool for pgx.
pgxpool implements a nearly identical interface to pgx connections.

- The `PostgresDB` is stateful because it holds a connection to the database (`pgxpool.Pool`). This dependency is injected into the repository to manage database operations.
- This pattern is useful for managing resources that have a lifecycle, like database connections.
*/
type PostgresDB struct {
	Dbpool *pgxpool.Pool

	postgresDBConfig *entity.PostgresDBConfig
}

func (p *PostgresDB) Connect(postgresDBConfig *entity.PostgresDBConfig) repository.DatabaseConnection {
	connString := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=%s pool_max_conns=%s",
		postgresDBConfig.Username, postgresDBConfig.Password,
		postgresDBConfig.Host, postgresDBConfig.Port,
		postgresDBConfig.Name, postgresDBConfig.SslMode,
		postgresDBConfig.PoolMaxConns,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Ctx(ctx).Error().Err(err).Msg(errMsgContextTimeout)
		}

		log.Error().Err(err).Msg("unable to create connection pool")
		panic(err)
	}

	var greeting string
	err = dbpool.QueryRow(ctx, "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Ctx(ctx).Error().Err(err).Msg(errMsgContextTimeout)
		}

		log.Error().Err(err).Msg("dbpool.QueryRow failed")
		panic(err)
	}

	log.Info().Msg("successfully connected to the Postgres database")
	return &PostgresDB{Dbpool: dbpool, postgresDBConfig: postgresDBConfig}
}

func (p *PostgresDB) Disconnect() {
	if p.Dbpool != nil {
		p.Dbpool.Close()
		log.Info().Msg("PostgreSQL database connection closed")
	}
}
