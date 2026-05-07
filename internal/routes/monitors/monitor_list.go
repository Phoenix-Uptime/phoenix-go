package monitors

import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MonitorListResponse struct {
	Monitors []MonitorResponse `json:"monitors"`
}

// @Summary List Monitors
// @Description Returns monitors owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} MonitorListResponse "monitor list"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors [get]
func ListMonitors(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	monitors, err := database.Client.Monitor.Query().
		Where(entmonitor.UserID(user.ID)).
		Order(entmonitor.ByCreatedAt(entsql.OrderDesc())).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list monitors")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list monitors",
		})
	}

	response := MonitorListResponse{
		Monitors: make([]MonitorResponse, 0, len(monitors)),
	}
	for _, monitor := range monitors {
		response.Monitors = append(response.Monitors, monitorResponse(monitor))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
