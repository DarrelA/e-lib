package config

import (
	"os"
	"strings"

	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	errMsgVarNotSet = "%s is not set"
)

type LoadEnvConfig interface {
	LoadServerConfig()
	LoadLogConfig()
	LoadPostgresConfig()
	LoadRedisConfig()
	LoadOAuth2Config()
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

func (e *EnvConfig) LoadLogConfig() {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel != "trace" && logLevel != "debug" &&
		logLevel != "info" && logLevel != "warn" &&
		logLevel != "error" && logLevel != "fatal" &&
		logLevel != "panic" {
		log.Error().Msgf("LOG_LEVEL is set to [%s]; only 'trace', 'debug', 'info', 'warn', 'error', 'fatal', 'panic' are accepted", logLevel)
	}

	// Whichever level is chosen,
	// all logs with a level greater than or equal to that level will be written.
	switch logLevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel) // Level -1
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // Level 0
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // Level 1
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel) // Level 2
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel) // Level 3
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel) // Level 4
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel) // Level 5
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel) // Level 1
	}

	log.Info().Msgf("logLevel is set to [%s] in the [%s] env", logLevel, e.AppEnv)
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

func (e *EnvConfig) LoadRedisConfig() {
	e.RedisDBConfig = &entity.RedisDBConfig{
		RedisAddr: checkEmptyEnvVar("REDIS_ADDR"),
	}
}

func (e *EnvConfig) LoadOAuth2Config() {
	// Google Cloud Console -> Credentials -> OAuth 2.0 Client IDs -> Authorized redirect URIs
	e.OAuth2Config = &entity.OAuth2Config{
		GoogleRedirectURL:  "http://localhost:" + e.Port + "/auth/google_callback",
		GoogleClientID:     checkEmptyEnvVar("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: checkEmptyEnvVar("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
}

func checkEmptyEnvVar(envVar string) string {
	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		log.Error().Msgf(errMsgVarNotSet, envVar)
	}
	return valueStr
}
