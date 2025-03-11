package config

import (
	"os"
	"strings"

	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/rs/zerolog/log"
)

const (
	errMsgVarNotSet = "%s is not set"
)

type LoadEnvConfig interface {
	LoadServerConfig()
	LoadPostgresConfig()
}

type EnvConfig struct {
	entity.EnvConfig
}

func NewEnvConfig() LoadEnvConfig {
	return &EnvConfig{}
}

func (e *EnvConfig) LoadServerConfig() {
	e.AppEnv = strings.ToLower(checkEmptyEnvVar("APP_ENV"))
	e.Port = checkEmptyEnvVar("APP_PORT")
	e.PathToSQLSchema = "./config/schema.elib.sql"
	e.PathToBooksJsonFile = "./testdata/json/" + e.AppEnv + ".books.json"
}

func (e *EnvConfig) LoadPostgresConfig() {
	e.PostgresDBConfig = &entity.PostgresDBConfig{
		Username:     checkEmptyEnvVar("POSTGRES_USER"),
		Password:     checkEmptyEnvVar("POSTGRES_PASSWORD"),
		Host:         checkEmptyEnvVar("POSTGRES_HOST"),
		Port:         checkEmptyEnvVar("POSTGRES_PORT"),
		Name:         checkEmptyEnvVar("POSTGRES_DB"),
		SslMode:      checkEmptyEnvVar("POSTGRES_SSLMODE"),
		PoolMaxConns: checkEmptyEnvVar("POSTGRES_POOL_MAX_CONNS"),
	}
}

func checkEmptyEnvVar(envVar string) string {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		log.Error().Msgf(errMsgVarNotSet, envVar)
	}
	return valueStr
}
