package incidents

import (
	"strconv"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type IncidentListResponse struct {
	Incidents []IncidentResponse `json:"incidents"`
}

// @Summary List Incidents
// @Description Returns incidents for monitors owned by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param status query string false "open, acknowledged, or resolved"
// @Param severity query string false "info, warning, or critical"
// @Param monitor_id query int false "Monitor ID"
// @Success 200 {object} IncidentListResponse "incident list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents [get]
func ListIncidents(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.Incident.Query().
		Where(entincident.HasMonitorWith(entmonitor.UserID(user.ID))).
		Order(entincident.ByStartedAt(entsql.OrderDesc()))

	if status := c.Query("status"); status != "" {
		incidentStatus := entincident.Status(status)
		if err := entincident.StatusValidator(incidentStatus); err != nil {
			return badRequest(c, "Invalid incident status")
		}
		query.Where(entincident.StatusEQ(incidentStatus))
	}
	if severity := c.Query("severity"); severity != "" {
		incidentSeverity := entincident.Severity(severity)
		if err := entincident.SeverityValidator(incidentSeverity); err != nil {
			return badRequest(c, "Invalid incident severity")
		}
		query.Where(entincident.SeverityEQ(incidentSeverity))
	}
	if rawMonitorID := c.Query("monitor_id"); rawMonitorID != "" {
		monitorID, err := strconv.Atoi(rawMonitorID)
		if err != nil || monitorID < 1 {
			return badRequest(c, "Invalid monitor id")
		}
		query.Where(entincident.MonitorID(monitorID))
	}

	incidents, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list incidents")
		return serverError(c, "Failed to list incidents")
	}

	response := IncidentListResponse{
		Incidents: make([]IncidentResponse, 0, len(incidents)),
	}
	for _, incident := range incidents {
		response.Incidents = append(response.Incidents, incidentResponse(incident))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
