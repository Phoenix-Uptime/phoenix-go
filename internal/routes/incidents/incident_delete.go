package incidents

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"strconv"
)

// @Summary Delete Incident
// @Description Deletes one incident for a monitor owned by the authenticated user.
// @Tags Incidents
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Incident ID"
// @Success 200 {object} routes.SuccessResponse "incident deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid incident id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "incident not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /incidents/{id} [delete]
func DeleteIncident(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid incident id",
		})
	}
	if _, err := userIncident(c, user.ID, id); err != nil {
		return incidentLookupError(c, err)
	}

	if err := database.Client.Incident.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete incident")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete incident",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Incident deleted",
	})
}
