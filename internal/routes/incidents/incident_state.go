package incidents

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entincident "github.com/Phoenix-Uptime/phoenix-go/ent/incident"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Acknowledge Incident
// @Description Marks one incident acknowledged.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Incident ID"
// @Success 200 {object} IncidentResponse "acknowledged incident"
// @Failure 400 {object} routes.ErrorResponse "invalid incident id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "incident not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents/{id}/acknowledge [post]
func AcknowledgeIncident(c fiber.Ctx) error {
	return setIncidentStatus(c, entincident.StatusAcknowledged)
}

// @Summary Resolve Incident
// @Description Marks one incident resolved by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Incident ID"
// @Success 200 {object} IncidentResponse "resolved incident"
// @Failure 400 {object} routes.ErrorResponse "invalid incident id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "incident not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents/{id}/resolve [post]
func ResolveIncident(c fiber.Ctx) error {
	return setIncidentStatus(c, entincident.StatusResolved)
}

func setIncidentStatus(c fiber.Ctx, status entincident.Status) error {
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

	update := database.Client.Incident.UpdateOneID(id).SetStatus(status)
	if status == entincident.StatusResolved {
		update.SetResolvedByID(user.ID).SetEndedAt(time.Now())
	} else {
		update.ClearResolvedByID().ClearEndedAt()
	}

	incident, err := update.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update incident status")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update incident status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(incidentResponse(incident))
}
