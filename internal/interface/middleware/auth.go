package middleware

import (
	"strconv"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

const (
	errMsgPleaseLoginAgain = "please login again"
)

type GetSessionByIDFunc func(sessionID string) (*entity.Session, *apperrors.RestErr)
type GetUserByIDFunc func(userID int64) (dto.UserDetail, *apperrors.RestErr)

type AuthMiddleware struct {
	getSessionData GetSessionByIDFunc
	getUser        GetUserByIDFunc
}

func NewAuthMiddleware(getSessionData GetSessionByIDFunc, getUser GetUserByIDFunc) *AuthMiddleware {
	return &AuthMiddleware{getSessionData, getUser}
}

func (m *AuthMiddleware) Authenticate(c *fiber.Ctx) error {
	sessionID := c.Cookies("session_id")
	if sessionID == "" {
		log.Error().Msg("session_id cookie not found")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": errMsgPleaseLoginAgain})
	}

	sessionData, err := m.getSessionData(sessionID)
	if err != nil {
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	if sessionData.UserID == "" {
		log.Error().Msg("sessionData.UserID is an empty string")
		restErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(restErr.Status).JSON(restErr)
	}

	userID, convErr := strconv.ParseInt(sessionData.UserID, 10, 64)
	if convErr != nil {
		log.Error().Err(convErr).Msg("error converting sessionData.UserID (string) to int64")
		restErr := apperrors.NewInternalServerError(apperrors.ErrMsgSomethingWentWrong)
		return c.Status(restErr.Status).JSON(restErr)
	}

	userDetail, err := m.getUser(userID)
	if err != nil {
		log.Error().Err(err).Msg("error getting user by ID")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized: invalid user"})
	}

	log.Debug().Msgf("Current user: %d: %s", userDetail.ID, userDetail.Name)
	c.Locals("userDetail", userDetail)
	return c.Next()
}
