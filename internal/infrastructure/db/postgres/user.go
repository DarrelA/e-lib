package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type UserRepository struct {
	dbpool *pgxpool.Pool
}

func NewUserRepository(dbpool *pgxpool.Pool) repository.UserRepository {
	return &UserRepository{dbpool}
}

func (ur UserRepository) GetUser(provider string, id string, email string) (int, *apperrors.RestErr) {
	var user_id int

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const queryGetUserFromAuth = "SELECT user_id FROM AuthProviders WHERE provider=$1 AND id=$2 AND email=$3;"
	err := ur.dbpool.QueryRow(ctx, queryGetUserFromAuth, provider, id, email).Scan(&user_id)

	if err != nil {
		if err == context.DeadlineExceeded {
			log.Ctx(ctx).Error().Msg(errMsgContextTimeout)
			return -1, apperrors.NewInternalServerError(errMsgContextTimeout)
		}

		if errors.Is(err, pgx.ErrNoRows) { // Use errors.Is for accurate error comparison
			log.Info().Msgf("no user id found for the provider %s", provider)
			return -1, nil // Return -1 and *nil* error to indicate user not found
		}

		log.Error().Err(err).Msg("")
		return -1, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	return user_id, nil // return user_id and *nil* for error.
}

func (ur UserRepository) SaveUser(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr) {
	newUser := &entity.User{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := ur.dbpool.Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToBeginTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				log.Error().Err(errRollback).Msg(errMsgFailedToRollbackTransaction)
			} else {
				log.Info().Msg(infoMsgRollbackTransactionSuccess)
			}
		} else {
			log.Info().Msg(infoMsgCommittedTransactionSuccess)
		}
	}()

	const queryInsertUsers = "INSERT INTO Users (name, email, created_at, updated_at) VALUES ($1, $2, NOW(), NOW()) returning id, name, email, created_at, updated_at"
	err = tx.QueryRow(ctx, queryInsertUsers, user.Name, user.Email).
		Scan(&newUser.ID, &newUser.Name, &newUser.Email, &newUser.CreatedAt, &newUser.UpdatedAt)

	if err != nil {
		log.Error().Err(err).Msg("error saving new user into Users table")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	const execInsertAuthProviders = `
		INSERT INTO AuthProviders (id, user_id, name, email, verified_email, provider, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = tx.Exec(ctx, execInsertAuthProviders, user.ID, newUser.ID, user.Name, user.Email, user.VerifiedEmail, provider)
	if err != nil {
		log.Error().Err(err).Msg("error saving new user into AuthProviders table")
		rErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Error().Err(rbErr).Msg(errMsgFailedToRollbackTransaction)
		}
		return nil, rErr
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Error().Err(err).Msg(errMsgFailedToCommitTransaction)
		return nil, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}
	err = nil // clear the err, so Rollback wont be executed.

	return newUser, nil
}
