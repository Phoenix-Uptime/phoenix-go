package api_keys

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Revoke API Key
// @Description Revokes one API key owned by the authenticated user.
// @Tags API Keys
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "API Key ID"
// @Success 200 {object} routes.SuccessResponse "API key revoked"
// @Failure 400 {object} routes.ErrorResponse "invalid API key id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "API key not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /api-keys/{id} [delete]
func DeleteAPIKey(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := apiKeyID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid API key id",
		})
	}
	if _, err := userAPIKey(c, user.ID, id); err != nil {
		return apiKeyLookupError(c, err)
	}

	if _, err := database.Client.APIKey.UpdateOneID(id).
		SetIsActive(false).
		Save(c); err != nil {
		log.Error().Err(err).Msg("Failed to revoke API key")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to revoke API key",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "API key revoked",
	})
}
