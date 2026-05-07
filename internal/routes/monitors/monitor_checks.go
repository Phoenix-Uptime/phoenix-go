package monitors

import (
	"strconv"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entmonitorcheck "github.com/Phoenix-Uptime/phoenix-go/ent/monitorcheck"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type MonitorCheckResponse struct {
	ID              int       `json:"id"`
	MonitorID       int       `json:"monitor_id"`
	Status          string    `json:"status"`
	CheckedAt       time.Time `json:"checked_at"`
	ResponseTimeMs  int       `json:"response_time_ms"`
	StatusCode      *int      `json:"status_code,omitempty"`
	Message         *string   `json:"message,omitempty"`
	Error           *string   `json:"error,omitempty"`
	RetryCount      int       `json:"retry_count"`
	DurationSeconds int       `json:"duration_seconds"`
	Important       bool      `json:"important"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type MonitorChecksResponse struct {
	Checks []MonitorCheckResponse `json:"checks"`
}

// @Summary List Monitor Checks
// @Description Returns recent check results for one monitor owned by the authenticated user.
// @Tags Monitors
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Monitor ID"
// @Param limit query int false "Maximum checks to return"
// @Success 200 {object} MonitorChecksResponse "monitor checks"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "monitor not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /monitors/{id}/checks [get]
func ListMonitorChecks(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := monitorID(c)
	if err != nil {
		return badRequest(c, "Invalid monitor id")
	}
	if _, err := userMonitor(c, user.ID, id); err != nil {
		return monitorLookupError(c, err)
	}

	limit := 100
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil || parsedLimit < 1 {
			return badRequest(c, "Invalid limit")
		}
		limit = parsedLimit
	}
	if limit > 500 {
		limit = 500
	}

	checks, err := database.Client.MonitorCheck.Query().
		Where(entmonitorcheck.MonitorID(id)).
		Order(entmonitorcheck.ByCheckedAt(entsql.OrderDesc())).
		Limit(limit).
		All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list monitor checks")
		return serverError(c, "Failed to list monitor checks")
	}

	response := MonitorChecksResponse{
		Checks: make([]MonitorCheckResponse, 0, len(checks)),
	}
	for _, check := range checks {
		response.Checks = append(response.Checks, MonitorCheckResponse{
			ID:              check.ID,
			MonitorID:       check.MonitorID,
			Status:          string(check.Status),
			CheckedAt:       check.CheckedAt,
			ResponseTimeMs:  check.ResponseTimeMs,
			StatusCode:      check.StatusCode,
			Message:         check.Message,
			Error:           check.Error,
			RetryCount:      check.RetryCount,
			DurationSeconds: check.DurationSeconds,
			Important:       check.Important,
			CreatedAt:       check.CreatedAt,
			UpdatedAt:       check.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
