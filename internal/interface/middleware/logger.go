package middleware

import (
	"strings"
	"time"

	"github.com/DarrelA/e-lib/internal/apperrors"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func Logger(c *fiber.Ctx) error {
	appEnv, ok := c.Locals("appEnv").(string)
	if !ok {
		err := apperrors.NewInternalServerError("appEnv not found or has incorrect type")
		log.Error().Err(err).Msg("")
		return c.Status(err.Status).JSON(err)
	}

	start := time.Now()
	err := c.Next()
	duration := time.Since(start)

	bodyStr := "No body"
	bodyBytes := c.Body()
	contentType := string(c.Request().Header.ContentType())
	if strings.Contains(contentType, "application/json") {
		bodyStr = string(bodyBytes) // Log as JSON string
	} else if len(bodyBytes) > 0 {
		bodyStr = "Non-JSON Body"
	}

	log.Info().
		Str("method", c.Method()).
		Str("url", c.OriginalURL()).
		Str("request_body", bodyStr).
		Int("status", c.Response().StatusCode()).
		Str("user_host", c.Get("Host")).
		Dur("response_time", duration).
		Int64("latency_ms", duration.Milliseconds()).
		Str("user_agent", c.Get("User-Agent")).
		Msgf("request is completed in [%s] env", appEnv)

	return err
}
