package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type ResetAPIKeyResponse struct {
	Status string `json:"status"`
	ApiKey string `json:"api_key"`
}

// @Summary Reset API Key
// @Description Resets the API key for the authenticated user and returns the new API key.
// @Tags Account
// @Accept json
// @Produce json
// @Param api_key header string false "API Key for user authentication (Header)"
// @Param api_key query string false "API Key for user authentication (Query)"
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} ResetAPIKeyResponse "new API key"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/reset-api-key [post]
func ResetAPIKey(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	// Generate a new API key
	newApiKey := uuid.New().String()

	user.ApiKey = newApiKey
	if err := models.DB.Save(&user).Error; err != nil {
		log.Error().Err(err).Msg("Failed to reset API key")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to reset API key",
		})
	}

	return c.Status(fiber.StatusOK).JSON(ResetAPIKeyResponse{
		Status: "success",
		ApiKey: newApiKey,
	})
}
