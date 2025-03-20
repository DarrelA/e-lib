package middleware

import (
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
type GetUserByIDFunc func(userID int64) (*dto.UserDetail, *apperrors.RestErr)

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

	user, err := m.getUser(sessionData.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Error getting user by ID")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized: invalid user"})
	}

	c.Locals("user", user)
	return c.Next()
}
