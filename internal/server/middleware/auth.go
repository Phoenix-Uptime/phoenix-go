package middleware

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entapikey "github.com/Phoenix-Uptime/phoenix-go/ent/apikey"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware is a Fiber middleware for API key
func AuthMiddleware(c fiber.Ctx) error {
	apiKey := c.Get("x-api-key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	// Return 401 if no authentication method is provided
	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "API key is required",
		})
	}

	key, err := database.Client.APIKey.Query().
		Where(
			entapikey.Key(apiKey),
			entapikey.IsActive(true),
			entapikey.Or(
				entapikey.ExpiresAtIsNil(),
				entapikey.ExpiresAtGT(time.Now()),
			),
		).
		Only(c)
	if err != nil {
		log.Error().Err(err).Msg("Invalid API key")
		if !ent.IsNotFound(err) {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Internal server error",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid API key",
		})
	}

	user, err := key.QueryUser().Only(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load API key user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Internal server error",
		})
	}

	if err := database.Client.APIKey.UpdateOneID(key.ID).
		SetLastUsedAt(time.Now()).
		Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to update API key last used timestamp")
	}

	// Attach user to the context for use in subsequent handlers
	c.Locals("user", user)

	// Proceed to the next middleware or handler
	return c.Next()
}
