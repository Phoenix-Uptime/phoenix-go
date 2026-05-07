package monitors

import (
	"strconv"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitorstat "github.com/Phoenix-Uptime/phoenix-go/ent/monitorstat"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MonitorStatResponse struct {
	ID                int       `json:"id"`
	MonitorID         int       `json:"monitor_id"`
	Period            string    `json:"period"`
	PeriodStart       time.Time `json:"period_start"`
	TotalChecks       int       `json:"total_checks"`
	UpChecks          int       `json:"up_checks"`
	DownChecks        int       `json:"down_checks"`
	MaintenanceChecks int       `json:"maintenance_checks"`
	UptimePercentage  float64   `json:"uptime_percentage"`
	AvgResponseTimeMs int       `json:"avg_response_time_ms"`
	MinResponseTimeMs *int      `json:"min_response_time_ms,omitempty"`
	MaxResponseTimeMs *int      `json:"max_response_time_ms,omitempty"`
	DowntimeSeconds   int       `json:"downtime_seconds"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type MonitorStatsResponse struct {
	Stats []MonitorStatResponse `json:"stats"`
}

// @Summary List Monitor Stats
// @Description Returns stats for one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Param period query string false "hour or day"
// @Param limit query int false "Maximum stat rows to return"
// @Success 200 {object} MonitorStatsResponse "monitor stats"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/stats [get]
func ListMonitorStats(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid monitor id",
		})
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	limit := 100
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid limit",
			})
		}
		limit = parsedLimit
	}
	if limit > 500 {
		limit = 500
	}

	query := database.Client.MonitorStat.Query().
		Where(entmonitorstat.MonitorID(id)).
		Order(entmonitorstat.ByPeriodStart(entsql.OrderDesc())).
		Limit(limit)

	if period := c.Query("period"); period != "" {
		statPeriod := entmonitorstat.Period(period)
		if err := entmonitorstat.PeriodValidator(statPeriod); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid period",
			})
		}
		query.Where(entmonitorstat.PeriodEQ(statPeriod))
	}

	stats, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list monitor stats")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list monitor stats",
		})
	}

	response := MonitorStatsResponse{
		Stats: make([]MonitorStatResponse, 0, len(stats)),
	}
	for _, stat := range stats {
		response.Stats = append(response.Stats, MonitorStatResponse{
			ID:                stat.ID,
			MonitorID:         stat.MonitorID,
			Period:            string(stat.Period),
			PeriodStart:       stat.PeriodStart,
			TotalChecks:       stat.TotalChecks,
			UpChecks:          stat.UpChecks,
			DownChecks:        stat.DownChecks,
			MaintenanceChecks: stat.MaintenanceChecks,
			UptimePercentage:  stat.UptimePercentage,
			AvgResponseTimeMs: stat.AvgResponseTimeMs,
			MinResponseTimeMs: stat.MinResponseTimeMs,
			MaxResponseTimeMs: stat.MaxResponseTimeMs,
			DowntimeSeconds:   stat.DowntimeSeconds,
			CreatedAt:         stat.CreatedAt,
			UpdatedAt:         stat.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
