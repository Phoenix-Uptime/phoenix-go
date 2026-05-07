package maintenance_windows

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Activate Maintenance Window
// @Description Marks one maintenance window active.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Maintenance Window ID"
// @Success 200 {object} MaintenanceWindowResponse "active maintenance window"
// @Failure 400 {object} routes.ErrorResponse "invalid maintenance window id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "maintenance window not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows/{id}/activate [post]
func ActivateMaintenanceWindow(c fiber.Ctx) error {
	return setMaintenanceWindowActive(c, true)
}

// @Summary Deactivate Maintenance Window
// @Description Marks one maintenance window inactive.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Maintenance Window ID"
// @Success 200 {object} MaintenanceWindowResponse "inactive maintenance window"
// @Failure 400 {object} routes.ErrorResponse "invalid maintenance window id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "maintenance window not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows/{id}/deactivate [post]
func DeactivateMaintenanceWindow(c fiber.Ctx) error {
	return setMaintenanceWindowActive(c, false)
}

func setMaintenanceWindowActive(c fiber.Ctx, isActive bool) error {
	user := c.Locals("user").(*ent.User)
	id, err := maintenanceWindowID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid maintenance window id",
		})
	}
	if _, err := userMaintenanceWindow(c, user.ID, id); err != nil {
		return maintenanceWindowLookupError(c, err)
	}

	if _, err := database.Client.MaintenanceWindow.UpdateOneID(id).SetIsActive(isActive).Save(c); err != nil {
		log.Error().Err(err).Msg("Failed to update maintenance window state")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update maintenance window state",
		})
	}

	window, err := userMaintenanceWindow(c, user.ID, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload maintenance window")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update maintenance window state",
		})
	}

	return c.Status(fiber.StatusOK).JSON(maintenanceWindowResponse(window))
}
