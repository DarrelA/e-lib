package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	sessionPrefix = "session:"
	sessionTTL    = 24 * time.Hour // Session Time To Live

	errMsgLoginAgain      = "please login again"
	errMsgSessionNotFound = "session not found"
)

type SessionRepository struct {
	redisClient *redis.Client
	ctx         context.Context
}

func NewSessionRepository(redisClient *redis.Client) repository.SessionRepository {
	ctx := context.Background()
	return &SessionRepository{redisClient, ctx}
}

func (sr SessionRepository) NewSession(userID int64) (string, *apperrors.RestErr) {
	sessionID := uuid.New().String()
	sessionKey := sessionPrefix + sessionID

	userIDString := strconv.FormatInt(userID, 10)       // Convert int64 to string
	createdAt := time.Now().Unix()                      // Get Unix timestamp
	createdAtString := strconv.FormatInt(createdAt, 10) // Convert Unix timestamp to string
	sessionData := map[string]interface{}{              // Use a map to pass values
		"userID":    userIDString,
		"createdAt": createdAtString,
	}

	err := sr.redisClient.HSet(sr.ctx, sessionKey, sessionData).Err()
	if err != nil {
		log.Error().Err(err).Msg("")
		return "", apperrors.NewInternalServerError("failed to store session in Redis")
	}

	err = sr.redisClient.Expire(sr.ctx, sessionKey, sessionTTL).Err()
	if err != nil {
		log.Warn().Msg(fmt.Sprintf("failed to set session expiration: %s", err))
		return "", apperrors.NewInternalServerError("failed to set session expiration")
	}

	return sessionID, nil
}

func (r *SessionRepository) GetSessionData(sessionID string) (*entity.Session, *apperrors.RestErr) {
	ctx := context.Background()
	key := "session:" + sessionID

	sessionDataMap, err := r.redisClient.HGetAll(ctx, key).Result() // Changed to HGetAll
	if err == redis.Nil {
		log.Error().Err(err).Msg(errMsgSessionNotFound)
		return nil, apperrors.NewInternalServerError(errMsgSessionNotFound)
	} else if err != nil {
		log.Error().Err(err).Msg("failed to retrieve session data")
		return nil, apperrors.NewInternalServerError(errMsgLoginAgain)
	}

	userID, ok := sessionDataMap["userID"]
	if !ok {
		log.Error().Msg("userID not found in session data")
		return nil, apperrors.NewInternalServerError(errMsgLoginAgain)
	}

	createdAt, ok := sessionDataMap["createdAt"]
	if !ok {
		log.Error().Msg("createdAt not found in session data")
		return nil, apperrors.NewInternalServerError(errMsgLoginAgain)
	}

	sessionData := &entity.Session{
		UserID:    userID,
		CreatedAt: createdAt,
	}

	return sessionData, nil
}
