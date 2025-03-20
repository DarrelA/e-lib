package redis

import (
	"context"
	"encoding/json"
	"fmt"
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
	sessionData := &entity.Session{UserID: userID, CreatedAt: time.Now()}

	values := map[string]interface{}{
		"userID":    sessionData.UserID,
		"createdAt": sessionData.CreatedAt,
	}

	err := sr.redisClient.HSet(sr.ctx, sessionKey, values).Err()
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

	sessionDataString, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError("session not found")
	} else if err != nil {
		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError("please login again")
	}

	var sessionData entity.Session
	err = json.Unmarshal([]byte(sessionDataString), &sessionData)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal session data")
		return nil, apperrors.NewInternalServerError("please login again")
	}

	return &sessionData, nil
}
