package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	repository "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

const (
	errMsgUserNotFound  = "user not found"
	errMsgSavingNewUser = "error saving new user into postgres"
)

type UserRepository struct {
	dbpool *pgxpool.Pool
}

func NewUserRepository(dbpool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{dbpool}
}

var (
	queryGetUserFromAuth = "SELECT user_id FROM AuthProviders WHERE provider=$1 AND id=$2 AND email=$3;"

	queryInsertUser = `
  INSERT INTO Users (name, email, created_at, updated_at)
  VALUES ($1, $2, NOW(), NOW())
  returning id, name, email, created_at, updated_at
`
)

func (ur UserRepository) GetUser(provider string, id string, email string) (int, *apperrors.RestErr) {
	var user_id int

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info().Msgf("GetUser: provider=%s, id=%s, email=%s", provider, id, email)

	err := ur.dbpool.QueryRow(ctx, queryGetUserFromAuth, provider, id, email).Scan(&user_id)

	if err != nil {
		if err == context.DeadlineExceeded {
			errMsg := fmt.Sprintf("GetUser: Timeout while retrieving user %s", provider)
			log.Ctx(ctx).Error().Msg(errMsg)
			return -1, apperrors.NewInternalServerError(errMsg)
		}

		if errors.Is(err, pgx.ErrNoRows) { // Use errors.Is for accurate error comparison
			log.Info().Msg("GetUser: No user found")
			return -1, nil // Return -1 and *nil* error to indicate user not found
		}

		errMsg := "GetUser: Database error retrieving user"
		log.Error().Err(err).Msg(errMsg)
		return -1, apperrors.NewInternalServerError(errMsg)
	}

	log.Info().Msgf("GetUser: User found with id=%d", user_id)
	return user_id, nil // return user_id and *nil* for error.
}

func (ur UserRepository) SaveUser(user *dto.GoogleOAuth2UserRes) (*entity.User, *apperrors.RestErr) {
	newUser := &entity.User{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ur.dbpool.QueryRow(ctx, queryInsertUser, user.Name, user.Email).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Ctx(ctx).Error().Msg("Timeout occurred while saving user.")
			return nil, apperrors.NewInternalServerError("Timeout occurred while saving user.")
		}

		log.Error().Err(err).Msg("")
		return nil, apperrors.NewInternalServerError(errMsgSavingNewUser)
	}

	return newUser, nil
}
