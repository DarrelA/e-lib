package services

import "github.com/gofiber/fiber/v2"

type GoogleOAuth2Service interface {
	Login(c *fiber.Ctx) error
	Callback(c *fiber.Ctx) error
}
