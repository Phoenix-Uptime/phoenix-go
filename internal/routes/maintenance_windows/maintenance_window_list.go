package maintenance_windows

import (
	"strconv"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmaintenancewindow "github.com/Phoenix-Uptime/phoenix-go/ent/maintenancewindow"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MaintenanceWindowListResponse struct {
	MaintenanceWindows []MaintenanceWindowResponse `json:"maintenance_windows"`
}

// @Summary List Maintenance Windows
// @Description Returns maintenance windows owned by the authenticated user.
// @Tags Maintenance Windows
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param is_active query bool false "Filter by active flag"
// @Param strategy query string false "manual, single, recurring, or cron"
// @Param monitor_id query int false "Monitor ID"
// @Success 200 {object} MaintenanceWindowListResponse "maintenance window list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /maintenance-windows [get]
func ListMaintenanceWindows(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.MaintenanceWindow.Query().
		Where(entmaintenancewindow.UserID(user.ID)).
		WithMonitors().
		Order(entmaintenancewindow.ByCreatedAt(entsql.OrderDesc()))

	if rawActive := c.Query("is_active"); rawActive != "" {
		isActive, err := strconv.ParseBool(rawActive)
		if err != nil {
			return badRequest(c, "Invalid is_active value")
		}
		query.Where(entmaintenancewindow.IsActive(isActive))
	}
	if rawStrategy := c.Query("strategy"); rawStrategy != "" {
		strategy := entmaintenancewindow.Strategy(rawStrategy)
		if err := entmaintenancewindow.StrategyValidator(strategy); err != nil {
			return badRequest(c, "Invalid maintenance window strategy")
		}
		query.Where(entmaintenancewindow.StrategyEQ(strategy))
	}
	if rawMonitorID := c.Query("monitor_id"); rawMonitorID != "" {
		monitorID, err := strconv.Atoi(rawMonitorID)
		if err != nil || monitorID < 1 {
			return badRequest(c, "Invalid monitor id")
		}
		query.Where(entmaintenancewindow.HasMonitorsWith(entmonitor.ID(monitorID), entmonitor.UserID(user.ID)))
	}

	windows, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list maintenance windows")
		return serverError(c, "Failed to list maintenance windows")
	}

	response := MaintenanceWindowListResponse{
		MaintenanceWindows: make([]MaintenanceWindowResponse, 0, len(windows)),
	}
	for _, window := range windows {
		response.MaintenanceWindows = append(response.MaintenanceWindows, maintenanceWindowResponse(window))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
