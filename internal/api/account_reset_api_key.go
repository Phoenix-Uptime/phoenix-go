package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
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
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} ResetAPIKeyResponse "new API key"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/reset-api-key [post]
func ResetAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	// Generate a new API key
	newApiKey := uuid.New().String()

	if err := database.Client.User.UpdateOneID(user.ID).
		SetAPIKey(newApiKey).
		Exec(c); err != nil {
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
