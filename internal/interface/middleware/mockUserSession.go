package middleware

import (
	"fmt"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/DarrelA/e-lib/internal/domain/entity"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type NewSessionFunc func(userID int64) (string, *apperrors.RestErr)
type SaveUserFunc func(user *dto.GoogleOAuth2UserRes, provider string) (*entity.User, *apperrors.RestErr)

type MockUserSessionMiddleware struct {
	newSessionFunc NewSessionFunc
	saveUserFunc   SaveUserFunc
	hasDummyUser   bool
}

func NewMockUserSessionMiddleware(newSessionFunc NewSessionFunc, saveUserFunc SaveUserFunc) *MockUserSessionMiddleware {
	hasDummyUser := false
	return &MockUserSessionMiddleware{newSessionFunc, saveUserFunc, hasDummyUser}
}

// Simulate user login for testing environment
func (m *MockUserSessionMiddleware) New(c *fiber.Ctx) error {
	appEnv, ok := c.Locals("appEnv").(string)
	if !ok {
		err := apperrors.NewInternalServerError("appEnv not found or has incorrect type")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	if appEnv == "test" {
		sessionID, _ := m.newSessionFunc(1)
		cookieValue := fmt.Sprintf("session_id=%s", sessionID)
		c.Request().Header.Set("Cookie", cookieValue)
		c.Request().Header.VisitAllCookie(func(key, value []byte) {
			log.Debug().Msgf("Cookie: %s=%s", string(key), string(value))
		})

		if !m.hasDummyUser {
			googleOAuth2UserRes := &dto.GoogleOAuth2UserRes{
				ID:            "dummy_user_id_from_google",
				Name:          "dummy_user1",
				Email:         "dummy_user1@email.com",
				VerifiedEmail: true,
			}
			m.saveUserFunc(googleOAuth2UserRes, "google")
			m.hasDummyUser = true
		}
	}

	return c.Next()
}
