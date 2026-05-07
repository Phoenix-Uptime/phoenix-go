package monitors

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Pause Monitor
// @Description Marks a monitor paused and inactive.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Success 200 {object} MonitorResponse "paused monitor"
// @Failure 400 {object} routes.ErrorResponse "invalid monitor id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/pause [post]
func PauseMonitor(c fiber.Ctx) error {
	return setMonitorActiveState(c, false, entmonitor.StatusPaused, "Failed to pause monitor")
}

// @Summary Resume Monitor
// @Description Marks a monitor active and pending its next check.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Success 200 {object} MonitorResponse "resumed monitor"
// @Failure 400 {object} routes.ErrorResponse "invalid monitor id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/resume [post]
func ResumeMonitor(c fiber.Ctx) error {
	return setMonitorActiveState(c, true, entmonitor.StatusPending, "Failed to resume monitor")
}

func setMonitorActiveState(c fiber.Ctx, isActive bool, status entmonitor.Status, failureMessage string) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor id",
		})
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	updated, err := database.Client.Monitor.UpdateOneID(id).
		SetIsActive(isActive).
		SetStatus(status).
		Save(c)
	if err != nil {
		log.Error().Err(err).Msg(failureMessage)
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: failureMessage,
		})
	}

	return c.Status(fiber.StatusOK).JSON(monitorResponse(updated))
}
