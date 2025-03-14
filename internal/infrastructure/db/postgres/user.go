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
	errMsgUserNotFound               = "user not found"
	errMsgSaveNewUserInUsers         = "error saving new user into Users table"
	errMsgSaveNewUserInAuthProviders = "error saving new user into AuthProviders table"
)

type UserRepository struct {
	dbpool *pgxpool.Pool
}

func NewUserRepository(dbpool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{dbpool}
}

var (
	queryGetUserFromAuth = "SELECT user_id FROM AuthProviders WHERE provider=$1 AND id=$2 AND email=$3;"

	queryInsertUsers = `
  INSERT INTO Users (name, email, created_at, updated_at)
  VALUES ($1, $2, NOW(), NOW())
  returning id, name, email, created_at, updated_at
	`

	execInsertAuthProviders = `
  INSERT INTO AuthProviders (id, user_id, name, email, verified_email, provider, created_at, updated_at)
  VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
)

func (ur UserRepository) GetUser(provider string, id string, email string) (int, *apperrors.RestErr) {
	var user_id int

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := ur.dbpool.QueryRow(ctx, queryGetUserFromAuth, provider, id, email).Scan(&user_id)

	if err != nil {
		if err == context.DeadlineExceeded {
			errMsg := fmt.Sprintf("GetUser: Timeout while retrieving user %s", provider)
			log.Ctx(ctx).Error().Msg(errMsg)
			return -1, apperrors.NewInternalServerError(errMsg)
		}

		if errors.Is(err, pgx.ErrNoRows) { // Use errors.Is for accurate error comparison
			log.Info().Msgf("GetUser: No user found for %s", provider)
			return -1, nil // Return -1 and *nil* error to indicate user not found
		}

		errMsg := "GetUser: Database error retrieving user"
		log.Error().Err(err).Msg(errMsg)
		return -1, apperrors.NewInternalServerError(errMsg)
	}

	log.Info().Msgf("GetUser: User found with id=%d for %s", user_id, provider)
	return user_id, nil // return user_id and *nil* for error.
}

func (ur UserRepository) SaveUser(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr) {
	newUser := &entity.User{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := ur.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg("Transaction rollback failed")
			} else {
				log.Info().Msg("Transaction rollback successfully")
			}
		} else {
			log.Info().Msg("Transaction committed successfully")
		}
	}()

	err = tx.QueryRow(ctx, queryInsertUsers, user.Name, user.Email).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)

	if err != nil {
		log.Error().Err(err).Msg(errMsgSaveNewUserInAuthProviders)
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg("Rollback failed during error handling")
		}
		return nil, rErr
	}

	_, err = tx.Exec(ctx, execInsertAuthProviders, user.ID, newUser.ID, user.Name, user.Email, user.VerifiedEmail, provider)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert auth provider")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg("Rollback failed during error handling")
		}
		return nil, rErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to commit transaction")
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	err = nil // clear the err, so Rollback wont be executed.

	return newUser, nil
}
