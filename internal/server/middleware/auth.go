package middleware

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware is a Fiber middleware for API key
func AuthMiddleware(c *fiber.Ctx) error {
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

	// Find user by API key
	var user models.User
	if err := models.DB.Where("api_key = ?", apiKey).First(&user).Error; err != nil {
		log.Error().Err(err).Msg("Invalid API key")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid API key",
		})
	}

	// Attach user to the context for use in subsequent handlers
	c.Locals("user", &user)

	// Proceed to the next middleware or handler
	return c.Next()
}
