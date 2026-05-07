package incidents

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	entstatuspage "github.com/Phoenix-Uptime/phoenix-go/ent/statuspage"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type IncidentResponse struct {
	ID           int        `json:"id"`
	MonitorID    int        `json:"monitor_id"`
	StatusPageID *int       `json:"status_page_id,omitempty"`
	ResolvedByID *int       `json:"resolved_by_id,omitempty"`
	Title        string     `json:"title"`
	Content      *string    `json:"content,omitempty"`
	Status       string     `json:"status"`
	Severity     string     `json:"severity"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	IsPinned     bool       `json:"is_pinned"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// @Summary Get Incident
// @Description Returns one incident for a monitor owned by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Incident ID"
// @Success 200 {object} IncidentResponse "incident"
// @Failure 400 {object} routes.ErrorResponse "invalid incident id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "incident not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents/{id} [get]
func GetIncident(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := incidentID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid incident id",
		})
	}

	incident, err := userIncident(c, user.ID, id)
	if err != nil {
		return incidentLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(incidentResponse(incident))
}

func incidentID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userIncident(c fiber.Ctx, userID int, id int) (*ent.Incident, error) {
	return database.Client.Incident.Query().
		Where(
			entincident.ID(id),
			entincident.HasMonitorWith(entmonitor.UserID(userID)),
		).
		Only(c)
}

func userOwnsMonitor(c fiber.Ctx, userID int, monitorID int) (bool, error) {
	return database.Client.Monitor.Query().
		Where(
			entmonitor.ID(monitorID),
			entmonitor.UserID(userID),
		).
		Exist(c)
}

func userOwnsStatusPage(c fiber.Ctx, userID int, statusPageID int) (bool, error) {
	return database.Client.StatusPage.Query().
		Where(
			entstatuspage.ID(statusPageID),
			entstatuspage.UserID(userID),
		).
		Exist(c)
}

func incidentLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Incident not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load incident")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load incident",
	})
}

func incidentResponse(incident *ent.Incident) IncidentResponse {
	return IncidentResponse{
		ID:           incident.ID,
		MonitorID:    incident.MonitorID,
		StatusPageID: incident.StatusPageID,
		ResolvedByID: incident.ResolvedByID,
		Title:        incident.Title,
		Content:      incident.Content,
		Status:       string(incident.Status),
		Severity:     string(incident.Severity),
		StartedAt:    incident.StartedAt,
		EndedAt:      incident.EndedAt,
		IsPinned:     incident.IsPinned,
		CreatedAt:    incident.CreatedAt,
		UpdatedAt:    incident.UpdatedAt,
	}
}
