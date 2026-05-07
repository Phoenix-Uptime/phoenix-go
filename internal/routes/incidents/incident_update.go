package incidents

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateIncidentRequest struct {
	MonitorID    *int       `json:"monitor_id,omitempty" validate:"omitempty,min=1"`
	StatusPageID *int       `json:"status_page_id,omitempty" validate:"omitempty,min=0"`
	Title        *string    `json:"title,omitempty" validate:"omitempty,min=1"`
	Content      *string    `json:"content,omitempty"`
	Status       *string    `json:"status,omitempty" validate:"omitempty,oneof=open acknowledged resolved"`
	Severity     *string    `json:"severity,omitempty" validate:"omitempty,oneof=info warning critical"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	IsPinned     *bool      `json:"is_pinned,omitempty"`
}

// @Summary Update Incident
// @Description Updates one incident for a monitor owned by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Incident ID"
// @Param data body UpdateIncidentRequest true "incident payload"
// @Success 200 {object} IncidentResponse "updated incident"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "incident not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents/{id} [patch]
func UpdateIncident(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := incidentID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid incident id",
		})
	}
	if _, err := userIncident(c, user.ID, id); err != nil {
		return incidentLookupError(c, err)
	}

	var req UpdateIncidentRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}
	if err := validator.New().Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	if req.MonitorID != nil {
		ok, err := userOwnsMonitor(c, user.ID, *req.MonitorID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate incident monitor")
			return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Failed to validate incident",
			})
		}
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid monitor id",
			})
		}
	}
	if req.StatusPageID != nil && *req.StatusPageID > 0 {
		ok, err := userOwnsStatusPage(c, user.ID, *req.StatusPageID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate incident status page")
			return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Failed to validate incident",
			})
		}
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid status page id",
			})
		}
	}

	update := database.Client.Incident.UpdateOneID(id)
	if req.MonitorID != nil {
		update.SetMonitorID(*req.MonitorID)
	}
	if req.StatusPageID != nil {
		if *req.StatusPageID == 0 {
			update.ClearStatusPageID()
		} else {
			update.SetStatusPageID(*req.StatusPageID)
		}
	}
	if req.Title != nil {
		update.SetTitle(*req.Title)
	}
	if req.Content != nil {
		if *req.Content == "" {
			update.ClearContent()
		} else {
			update.SetContent(*req.Content)
		}
	}
	if req.Severity != nil {
		update.SetSeverity(entincident.Severity(*req.Severity))
	}
	if req.StartedAt != nil {
		update.SetStartedAt(*req.StartedAt)
	}
	if req.EndedAt != nil {
		update.SetEndedAt(*req.EndedAt)
	}
	if req.IsPinned != nil {
		update.SetIsPinned(*req.IsPinned)
	}
	if req.Status != nil {
		status := entincident.Status(*req.Status)
		update.SetStatus(status)
		if status == entincident.StatusResolved {
			update.SetResolvedByID(user.ID)
			if req.EndedAt == nil {
				update.SetEndedAt(time.Now())
			}
		} else {
			update.ClearResolvedByID()
			update.ClearEndedAt()
		}
	}

	incident, err := update.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update incident")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update incident",
		})
	}

	return c.Status(fiber.StatusOK).JSON(incidentResponse(incident))
}
