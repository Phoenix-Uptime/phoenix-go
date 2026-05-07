package monitors

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Delete Monitor
// @Description Deletes one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Success 200 {object} routes.SuccessResponse "monitor deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid monitor id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id} [delete]
func DeleteMonitor(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return badRequest(c, "Invalid monitor id")
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	if err := database.Client.Monitor.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete monitor")
		return serverError(c, "Failed to delete monitor")
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Monitor deleted",
	})
}
