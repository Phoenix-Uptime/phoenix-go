package account

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entapikey "github.com/Phoenix-Uptime/phoenix-go/ent/apikey"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
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
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /account/reset-api-key [post]
func ResetAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	// Generate a new API key
	newApiKey := uuid.New().String()

	tx, err := database.Client.Tx(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start API key reset transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to reset API key",
		})
	}

	if err := tx.APIKey.Update().
		Where(entapikey.UserID(user.ID), entapikey.IsActive(true)).
		SetIsActive(false).
		Exec(c); err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to reset API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to reset API key",
		})
	}

	if _, err := tx.APIKey.Create().
		SetUserID(user.ID).
		SetName("Default").
		SetKey(newApiKey).
		Save(c); err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to create reset API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to reset API key",
		})
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit API key reset transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to reset API key",
		})
	}

	return c.Status(fiber.StatusOK).JSON(ResetAPIKeyResponse{
		Status: "success",
		ApiKey: newApiKey,
	})
}
