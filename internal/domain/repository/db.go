package repository

import "github.com/DarrelA/e-lib/internal/domain/entity"

type RDBMS interface {
	ConnectToPostgres(postgresDBConfig *entity.PostgresDBConfig) RDBMS
	DisconnectFromPostgres()
}
