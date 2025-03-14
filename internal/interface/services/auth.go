package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/application/services"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	repository "github.com/DarrelA/e-lib/internal/domain/repository/postgres"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

	errMsgPleaseLoginAgain = "please login again."
)

type GoogleOAuth2 struct {
	googleLoginConfig oauth2.Config
	userPGDB          repository.UserRepository
}

func NewGoogleOAuth2(
	OAuth2Config *entity.OAuth2Config, userPGDB repository.UserRepository,
) services.GoogleOAuth2Service {
	googleLoginConfig := oauth2.Config{
		RedirectURL:  OAuth2Config.GoogleRedirectURL,
		ClientID:     OAuth2Config.GoogleClientID,
		ClientSecret: OAuth2Config.GoogleClientSecret,
		Scopes:       OAuth2Config.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleOAuth2{googleLoginConfig, userPGDB}
}

func (oa GoogleOAuth2) Login(c *fiber.Ctx) error {
	url := oa.googleLoginConfig.AuthCodeURL("randomstate")
	return c.Redirect(url, fiber.StatusSeeOther)
}

func (oa GoogleOAuth2) Callback(c *fiber.Ctx) error {
	state := c.Query("state")
	if state != "randomstate" {
		err := apperrors.NewBadRequestError(errMsgPleaseLoginAgain)
		return c.Status(err.Status).JSON(err)
	}

	code := c.Query("code")
	token, err := oa.googleLoginConfig.Exchange(context.Background(), code)
	if err != nil {
		err := apperrors.NewBadRequestError(errMsgPleaseLoginAgain)
		return c.Status(err.Status).JSON(err)
	}

	resp, err := http.Get(googleUserInfoEndpoint + token.AccessToken)
	if err != nil {
		log.Error().Err(err).Msg("google oauth2 error")
		err := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(err.Status).JSON(err)
	}

	userInfo, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("google oauth2 error")
		err := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(err.Status).JSON(err)
	}

	user := &dto.GoogleOAuth2UserRes{}
	if err := json.Unmarshal(userInfo, &user); err != nil {
		log.Error().Err(err).Msg("")
		err := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(err.Status).JSON(err)
	}

	var wg sync.WaitGroup
	wg.Add(1) // Increment the counter

	var dbErrFromGoroutine *apperrors.RestErr

	go func() {
		defer wg.Done() // Decrement the counter when the goroutine finishes
		dbErr := oa.SaveUserToRDBMS(user)
		if dbErr != nil {
			log.Error().Err(dbErr).Msg("Error saving user to RDBMS in goroutine")
			dbErrFromGoroutine = dbErr // Copy the error to the shared variable!
		}
	}()

	wg.Wait()

	if dbErrFromGoroutine != nil {
		log.Error().Err(dbErrFromGoroutine).Msg("Error saving user to RDBMS after waitgroup")
		return c.Status(fiber.StatusInternalServerError).JSON(dbErrFromGoroutine)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": user})
}

func (oa GoogleOAuth2) SaveUserToRDBMS(user *dto.GoogleOAuth2UserRes) *apperrors.RestErr {
	user_id, dbErr := oa.userPGDB.GetUser("google", user.ID, user.Email)
	if dbErr != nil {
		return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	// No user_id found in User table
	if user_id == -1 {
		user, dbErr := oa.userPGDB.SaveUser(user, "google")
		if dbErr != nil {
			return apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		}

		log.Info().Msgf("User %s has joined the e-Lib!", user.Name)
	}

	return nil
}
