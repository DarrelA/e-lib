package redis

import (
	"context"
	"time"

	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	errMsgConnectingToDB      = "error connecting to the Redis database"
	errMsgDisconnectingFromDB = "error closing Redis database"
)

type RedisDB struct {
	redisDBConfig *entity.RedisDBConfig
	redisClient   *redis.Client
	ctx           context.Context
}

func (r *RedisDB) Connect(redisDBConfig *entity.RedisDBConfig) repository.DatabaseConnection {
	// Create a top level context
	ctx := context.Background()

	log.Info().Msgf("%s", redisDBConfig.RedisAddr)

	redisClient := redis.NewClient(&redis.Options{Addr: redisDBConfig.RedisAddr})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Error().Err(err).Msg(errMsgConnectingToDB)
		panic(err)
	}

	log.Info().Msg("successfully connected to the Redis database")
	return &RedisDB{redisDBConfig, redisClient, ctx}
}

func (r *RedisDB) Disconnect() {
	if r != nil {
		_, cancel := context.WithTimeout(r.ctx, 10*time.Second)
		defer cancel()

		err := r.redisClient.Close()
		if err != nil {
			log.Error().Err(err).Msg(errMsgDisconnectingFromDB)
		} else {
			log.Info().Msg("Redis database connection closed")
		}
	}
}
