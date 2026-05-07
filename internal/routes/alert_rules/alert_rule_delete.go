package alert_rules

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// @Summary Delete Alert Rule
// @Description Deletes one alert rule owned by the authenticated user.
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Alert Rule ID"
// @Success 200 {object} routes.SuccessResponse "alert rule deleted"
// @Failure 400 {object} routes.ErrorResponse "invalid alert rule id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "alert rule not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /alert-rules/{id} [delete]
func DeleteAlertRule(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := alertRuleID(c)
	if err != nil {
		return badRequest(c, "Invalid alert rule id")
	}
	if _, err := userAlertRule(c, user.ID, id); err != nil {
		return alertRuleLookupError(c, err)
	}

	if err := database.Client.AlertRule.DeleteOneID(id).Exec(c); err != nil {
		log.Error().Err(err).Msg("Failed to delete alert rule")
		return serverError(c, "Failed to delete alert rule")
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
		Status:  "success",
		Message: "Alert rule deleted",
	})
}
