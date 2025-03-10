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
	"github.com/DarrelA/e-lib/internal/infrastructure/db/filedb"
	"github.com/DarrelA/e-lib/internal/infrastructure/db/postgres"
	logger "github.com/DarrelA/e-lib/internal/infrastructure/logger/zerolog"
	interfaceSvc "github.com/DarrelA/e-lib/internal/interface/services"
	"github.com/DarrelA/e-lib/internal/interface/transport/rest"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	logFilePath         = "app.log"
	pathToBooksJsonFile = "./testdata/json/books.json"
)

func main() {
	logFile := logger.CreateAppLog(logFilePath)
	logger.NewZeroLogger(logFile)
	config := initializeEnv()
	postgresConn := initializeDatabases(config)

	// Use `WaitGroup` when you just need to wait for tasks to complete without exchanging data.
	// Use channels when you need to signal task completion and possibly exchange data.
	var wg sync.WaitGroup
	wg.Add(1)
	appServiceInstance := initializeServer(&wg, config)

	wg.Wait()
	waitForShutdown(appServiceInstance, postgresConn)
	logFile.Close()
	os.Exit(0)
}

func initializeEnv() *config.EnvConfig {
	envConfig := config.NewEnvConfig()
	envConfig.LoadServerConfig()
	envConfig.LoadPostgresConfig()
	config, ok := envConfig.(*config.EnvConfig)
	if !ok {
		log.Error().Msg("failed to load environment configuration")
	}

	return config
}

func initializeDatabases(config *config.EnvConfig) repository.RDBMS {
	postgresDB := &postgres.PostgresDB{}
	postgresConnection := postgresDB.ConnectToPostgres(config.PostgresDBConfig)
	postgresDBInstance := postgresConnection.(*postgres.PostgresDB) // Type assert postgresDB to *postgres.PostgresDB
	seedRepository := postgres.NewRepository(postgresDBInstance.Dbpool)
	seedRepository.SeedBooks(pathToBooksJsonFile)
	return postgresConnection
}

func initializeServer(wg *sync.WaitGroup, config *config.EnvConfig) *fiber.App {
	defer wg.Done()
	user := getDummyUserData()

	jsonFileService := filedb.NewJsonFileService()
	books := jsonFileService.LoadBooksJsonData()

	bookService := interfaceSvc.NewBookService(books)
	loanService := interfaceSvc.NewLoanService(user, bookService, jsonFileService)
	appInstance := rest.NewRouter(bookService, loanService)

	go func() {
		rest.StartServer(appInstance, config.Port)
	}()
	return appInstance
}

func waitForShutdown(appServiceInstance *fiber.App, postgresConn repository.RDBMS) {
	sigChan := make(chan os.Signal, 1) // Create a channel to listen for OS signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan // Block until a signal is received
	log.Debug().Msg("received termination signal, shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	if err := appServiceInstance.ShutdownWithContext(ctx); err != nil {
		log.Err(err).Msg("failed to gracefully shutdown the server")
	}

	cancel()
	log.Info().Msg("app instance has shutdown")

	postgresConn.DisconnectFromPostgres()
}

func getDummyUserData() entity.User {
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
	userDetail := entity.User{
		ID:        1,
		Name:      myInfoResponse.Name.Value,
		Email:     myInfoResponse.Email.Value,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	return userDetail
}
