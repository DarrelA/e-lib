package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DarrelA/e-lib/config"
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	pgdb "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/postgres"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/redis"
	logger "github.com/DarrelA/e-lib/internal/infrastructure/logger/zerolog"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/DarrelA/e-lib/internal/interface/transport/rest"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	logFilePath = "./config/app.log"
)

func main() {
	logFile := logger.CreateAppLog(logFilePath)
	logger.NewZeroLogger(logFile)
	config := initializeEnv()
	user, redisConn, postgresConn, postgresDBInstance,
		userRepository, bookRepository, loanRepository := initializeDatabases(config)

	// Use `WaitGroup` when you just need to wait for tasks to complete without exchanging data.
	// Use channels when you need to signal task completion and possibly exchange data.
	var wg sync.WaitGroup
	appInstance := initializeServer(&wg, user, config, postgresDBInstance,
		userRepository, bookRepository, loanRepository)

	wg.Wait()

	waitForShutdown(appInstance, redisConn, postgresConn)
	log.Info().Msg("Exiting...")
	logFile.Close()
	os.Exit(0)
}

func initializeEnv() *config.EnvConfig {
	envConfig := config.NewEnvConfig()
	envConfig.LoadRedisConfig()
	envConfig.LoadServerConfig()
	envConfig.LoadOAuth2Config()
	envConfig.LoadPostgresConfig()
	config, ok := envConfig.(*config.EnvConfig)
	if !ok {
		log.Error().Msg("failed to load environment configuration")
	}

	return config
}

func initializeDatabases(config *config.EnvConfig) (
	*entity.User, repository.InMemoryDB, repository.RDBMS,
	*postgres.PostgresDB, pgdb.UserRepository, pgdb.BookRepository, pgdb.LoanRepository,
) {
	user := getDummyUserData()
	redisDB := &redis.RedisDB{}
	redisConnection := redisDB.ConnectToRedis(config.RedisDBConfig)

	postgresDB := &postgres.PostgresDB{}
	postgresConnection := postgresDB.ConnectToPostgres(config.PostgresDBConfig)
	postgresDBInstance := postgresConnection.(*postgres.PostgresDB) // Type assert postgresDB to *postgres.PostgresDB

	seedRepository := postgres.NewRepository(config, postgresDBInstance.Dbpool, user)
	seedRepository.SeedBooks()

	userRepository := postgres.NewUserRepository(postgresDBInstance.Dbpool)
	bookRepository := postgres.NewBookRepository(postgresDBInstance.Dbpool)
	loanRepository := postgres.NewLoanRepository(postgresDBInstance.Dbpool)

	return user, redisConnection, postgresConnection, postgresDBInstance,
		userRepository, bookRepository, loanRepository
}

func initializeServer(
	wg *sync.WaitGroup, user *entity.User, config *config.EnvConfig,
	postgresDBInstance *postgres.PostgresDB,
	userRepository pgdb.UserRepository,
	bookRepository pgdb.BookRepository,
	loanRepository pgdb.LoanRepository,
) *fiber.App {

	wg.Add(1)
	defer wg.Done()

	googleOAuth2Service := interfaceSvc.NewGoogleOAuth2(config.OAuth2Config, userRepository)
	bookService := interfaceSvc.NewBookService(bookRepository)
	loanService := interfaceSvc.NewLoanService(*user, bookRepository, loanRepository)
	appInstance := rest.NewRouter(config, googleOAuth2Service, postgresDBInstance, bookService, loanService)

	go func() {
		rest.StartServer(appInstance, config.Port)
	}()
	return appInstance
}

func waitForShutdown(appInstance *fiber.App, redisConn repository.InMemoryDB, postgresConn repository.RDBMS) {
	sigChan := make(chan os.Signal, 1) // Create a channel to listen for OS signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan // Block until a signal is received
	log.Debug().Msg("received termination signal, shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := appInstance.ShutdownWithContext(ctx); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown the server")
	}

	cancel()
	log.Info().Msg("app instance has shutdown")

	redisConn.DisconnectFromRedis()
	postgresConn.DisconnectFromPostgres()
}

func getDummyUserData() *entity.User {
	url := "https://sandbox.api.myinfo.gov.sg/com/v4/person-sample/S9812381D"
	resp, err := http.Get(url)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	var myInfoResponse entity.MyInfoResponse
	err = json.Unmarshal(body, &myInfoResponse)
	if err != nil {
		log.Error().Msgf(apperrors.ErrMsgSomethingWentWrong)
	}

	currentTime := time.Now()
	userDetail := &entity.User{
		ID:        1,
		Name:      myInfoResponse.Name.Value,
		Email:     myInfoResponse.Email.Value,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	return userDetail
}
