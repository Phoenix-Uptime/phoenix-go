package public_status_pages

import (
	"strconv"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type PublicStatusPageIncidentResponse struct {
	ID        int        `json:"id"`
	MonitorID int        `json:"monitor_id"`
	Title     string     `json:"title"`
	Content   *string    `json:"content,omitempty"`
	Status    string     `json:"status"`
	Severity  string     `json:"severity"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
	IsPinned  bool       `json:"is_pinned"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type PublicStatusPageIncidentsResponse struct {
	Incidents []PublicStatusPageIncidentResponse `json:"incidents"`
}

// @Summary List Public Status Page Incidents
// @Description Returns incidents for a public status page by slug without API key authentication.
// @Tags Public Status Pages
// @Accept json
// @Produce json
// @Param slug path string true "Status Page Slug"
// @Param password query string false "Status page password when protected"
// @Param status query string false "open, acknowledged, or resolved"
// @Param limit query int false "Maximum incidents to return"
// @Success 200 {object} PublicStatusPageIncidentsResponse "public incident list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "password required or invalid"
// @Failure 404 {object} routes.ErrorResponse "status page not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /status-pages/public/{slug}/incidents [get]
func ListPublicStatusPageIncidents(c fiber.Ctx) error {
	page, err := publicStatusPage(c, c.Params("slug"))
	if err != nil {
		return publicStatusPageLookupError(c, err)
	}
	authorized, err := authorizePublicStatusPage(c, page)
	if err != nil {
		return err
	}
	if !authorized {
		return c.Status(fiber.StatusUnauthorized).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Status page password is required",
		})
	}

	monitorIDs := make([]int, 0, len(page.Edges.StatusPageMonitors))
	for _, assignment := range page.Edges.StatusPageMonitors {
		monitorIDs = append(monitorIDs, assignment.MonitorID)
	}
	query := database.Client.Incident.Query().
		Order(
			entincident.ByIsPinned(entsql.OrderDesc()),
			entincident.ByStartedAt(entsql.OrderDesc()),
		)
	if len(monitorIDs) == 0 {
		query.Where(entincident.StatusPageID(page.ID))
	} else {
		query.Where(entincident.Or(
			entincident.StatusPageID(page.ID),
			entincident.MonitorIDIn(monitorIDs...),
		))
	}

	if status := c.Query("status"); status != "" {
		incidentStatus := entincident.Status(status)
		if err := entincident.StatusValidator(incidentStatus); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid incident status",
			})
		}
		query.Where(entincident.StatusEQ(incidentStatus))
	}

	limit := 50
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
	if limit > 200 {
		limit = 200
	}
	query.Limit(limit)

	incidents, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list public status page incidents")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list incidents",
		})
	}

	response := PublicStatusPageIncidentsResponse{
		Incidents: make([]PublicStatusPageIncidentResponse, 0, len(incidents)),
	}
	for _, incident := range incidents {
		response.Incidents = append(response.Incidents, publicIncidentResponse(incident))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func publicIncidentResponse(incident *ent.Incident) PublicStatusPageIncidentResponse {
	return PublicStatusPageIncidentResponse{
		ID:        incident.ID,
		MonitorID: incident.MonitorID,
		Title:     incident.Title,
		Content:   incident.Content,
		Status:    string(incident.Status),
		Severity:  string(incident.Severity),
		StartedAt: incident.StartedAt,
		EndedAt:   incident.EndedAt,
		IsPinned:  incident.IsPinned,
		CreatedAt: incident.CreatedAt,
		UpdatedAt: incident.UpdatedAt,
	}
}
