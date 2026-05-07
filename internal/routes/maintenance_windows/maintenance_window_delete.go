package maintenance_windows

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"strconv"
)

// @Summary Delete Maintenance Window
// @Description Deletes one maintenance window owned by the authenticated user.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Maintenance Window ID"
// @Success 200 {object} routes.SuccessResponse "maintenance window deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid maintenance window id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "maintenance window not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows/{id} [delete]
func DeleteMaintenanceWindow(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid maintenance window id",
		})
	}
	if _, err := userMaintenanceWindow(c, user.ID, id); err != nil {
		return maintenanceWindowLookupError(c, err)
	}

	if err := database.Client.MaintenanceWindow.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete maintenance window")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete maintenance window",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Maintenance window deleted",
	})
}
