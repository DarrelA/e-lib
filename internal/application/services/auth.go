package services

import (
	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/DarrelA/e-lib/internal/application/dto"
	"github.com/gofiber/fiber/v2"
)

type GoogleOAuth2Service interface {
	Login(c *fiber.Ctx) error
	Callback(c *fiber.Ctx) error
	SaveUserToRDBMS(user *dto.GoogleOAuth2UserRes) (int64, *apperrors.RestErr)
}
