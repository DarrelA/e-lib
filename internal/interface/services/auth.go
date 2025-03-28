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
	"github.com/DarrelA/e-lib/internal/domain/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleUserInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

	errMsgPleaseLoginAgain = "please login again"
	errMsgGoogleOAuth2     = "google oauth2 error"
)

type GoogleOAuth2 struct {
	googleLoginConfig oauth2.Config
	userPGDB          repository.UserRepository
	sessionRedis      repository.SessionRepository
}

func NewGoogleOAuth2(OAuth2Config *entity.OAuth2Config,
	userPGDB repository.UserRepository, sessionRedis repository.SessionRepository,
) services.GoogleOAuth2Service {
	googleLoginConfig := oauth2.Config{
		RedirectURL:  OAuth2Config.GoogleRedirectURL,
		ClientID:     OAuth2Config.GoogleClientID,
		ClientSecret: OAuth2Config.GoogleClientSecret,
		Scopes:       OAuth2Config.Scopes,
		Endpoint:     google.Endpoint,
	}

	return &GoogleOAuth2{googleLoginConfig, userPGDB, sessionRedis}
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
		log.Error().Err(err).Msg(errMsgGoogleOAuth2)
		err := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(err.Status).JSON(err)
	}

	userInfo, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg(errMsgGoogleOAuth2)
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
	var libUserID int64

	go func() {
		defer wg.Done() // Decrement the counter when the goroutine finishes
		newLibUserID, dbErr := oa.SaveUserToRDBMS(user)
		libUserID = newLibUserID

		if dbErr != nil {
			log.Error().Err(dbErr).Msg("failed to save user into RDBMS in goroutine")
			dbErrFromGoroutine = dbErr // Copy the error to the shared variable!
		}
	}()

	wg.Wait()

	if dbErrFromGoroutine != nil {
		log.Error().Err(dbErrFromGoroutine).Msg("failed to save user into RDBMS after waitgroup")
		return c.Status(fiber.StatusInternalServerError).JSON(dbErrFromGoroutine)
	}

	sessionID, sessionErr := oa.sessionRedis.NewSession(libUserID)
	if sessionErr != nil {
		return sessionErr
	}

	cookie := fiber.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
	}
	c.Cookie(&cookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"user": user})
}

func (oa GoogleOAuth2) SaveUserToRDBMS(user *dto.GoogleOAuth2UserRes) (int64, *apperrors.RestErr) {
	user_id, dbErr := oa.userPGDB.GetUserID("google", user.ID, user.Email)
	if dbErr != nil {
		return -1, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
	}

	libUser := &entity.User{ID: int64(user_id)}

	// No user_id found in User table
	if user_id == -1 {
		libUser, dbErr = oa.userPGDB.SaveUser(user, "google")
		if dbErr != nil {
			return -1, apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		}

		log.Info().Msgf("user %s has joined the e-Lib using %s!", libUser.Name, libUser.Email)
	}

	return libUser.ID, nil
}
