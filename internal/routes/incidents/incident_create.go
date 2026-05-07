package incidents

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateIncidentRequest struct {
	MonitorID    int        `json:"monitor_id" validate:"required,min=1"`
	StatusPageID *int       `json:"status_page_id,omitempty" validate:"omitempty,min=1"`
	Title        string     `json:"title" validate:"required,min=1"`
	Content      *string    `json:"content,omitempty"`
	Status       string     `json:"status,omitempty" validate:"omitempty,oneof=open acknowledged resolved"`
	Severity     string     `json:"severity,omitempty" validate:"omitempty,oneof=info warning critical"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	IsPinned     *bool      `json:"is_pinned,omitempty"`
}

// @Summary Create Incident
// @Description Creates an incident for a monitor owned by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateIncidentRequest true "incident payload"
// @Success 201 {object} IncidentResponse "created incident"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents [post]
func CreateIncident(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateIncidentRequest
	if err := c.Bind().Body(&req); err != nil {
		return badRequest(c, "Invalid request payload")
	}
	if err := validator.New().Struct(&req); err != nil {
		return badRequest(c, "Invalid input: "+err.Error())
	}

	ok, err := userOwnsMonitor(c, user.ID, req.MonitorID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate incident monitor")
		return serverError(c, "Failed to validate incident")
	}
	if !ok {
		return badRequest(c, "Invalid monitor id")
	}
	if req.StatusPageID != nil {
		ok, err := userOwnsStatusPage(c, user.ID, *req.StatusPageID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate incident status page")
			return serverError(c, "Failed to validate incident")
		}
		if !ok {
			return badRequest(c, "Invalid status page id")
		}
	}

	status := entincident.StatusOpen
	if req.Status != "" {
		status = entincident.Status(req.Status)
	}
	create := database.Client.Incident.Create().
		SetMonitorID(req.MonitorID).
		SetTitle(req.Title).
		SetStatus(status)
	if req.StatusPageID != nil {
		create.SetStatusPageID(*req.StatusPageID)
	}
	if req.Content != nil {
		create.SetContent(*req.Content)
	}
	if req.Severity != "" {
		create.SetSeverity(entincident.Severity(req.Severity))
	}
	if req.StartedAt != nil {
		create.SetStartedAt(*req.StartedAt)
	}
	if req.EndedAt != nil {
		create.SetEndedAt(*req.EndedAt)
	}
	if req.IsPinned != nil {
		create.SetIsPinned(*req.IsPinned)
	}
	if status == entincident.StatusResolved {
		create.SetResolvedByID(user.ID)
		if req.EndedAt == nil {
			create.SetEndedAt(time.Now())
		}
	}

	incident, err := create.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create incident")
		return serverError(c, "Failed to create incident")
	}

	return c.Status(fiber.StatusCreated).JSON(incidentResponse(incident))
}
