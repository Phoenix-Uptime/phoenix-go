package alert_rules

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entalertrule "github.com/Phoenix-Uptime/phoenix-go/ent/alertrule"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateAlertRuleRequest struct {
	Name                   *string `json:"name,omitempty" validate:"omitempty,min=1"`
	Description            *string `json:"description,omitempty"`
	Event                  *string `json:"event,omitempty" validate:"omitempty,oneof=down recovered degraded certificate_expiring"`
	Scope                  *string `json:"scope,omitempty" validate:"omitempty,oneof=all tags monitors"`
	IsActive               *bool   `json:"is_active,omitempty"`
	ResendIntervalSeconds  *int    `json:"resend_interval_seconds,omitempty" validate:"omitempty,min=0"`
	TagIDs                 *[]int  `json:"tag_ids,omitempty"`
	MonitorIDs             *[]int  `json:"monitor_ids,omitempty"`
	NotificationChannelIDs *[]int  `json:"notification_channel_ids,omitempty"`
}

// @Summary Update Alert Rule
// @Description Updates one alert rule owned by the authenticated user.
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Alert Rule ID"
// @Param data body UpdateAlertRuleRequest true "alert rule payload"
// @Success 200 {object} AlertRuleResponse "updated alert rule"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "alert rule not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /alert-rules/{id} [patch]
func UpdateAlertRule(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := alertRuleID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid alert rule id",
		})
	}

	current, err := userAlertRule(c, user.ID, id)
	if err != nil {
		return alertRuleLookupError(c, err)
	}

	var req UpdateAlertRuleRequest
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

	scope := current.Scope
	if req.Scope != nil {
		scope = entalertrule.Scope(*req.Scope)
	}
	tagIDs := tagIDs(current.Edges.Tags)
	if req.TagIDs != nil {
		tagIDs = *req.TagIDs
	}
	monitorIDs := monitorIDs(current.Edges.Monitors)
	if req.MonitorIDs != nil {
		monitorIDs = *req.MonitorIDs
	}
	channelIDs := notificationChannelIDs(current.Edges.NotificationChannels)
	if req.NotificationChannelIDs != nil {
		channelIDs = *req.NotificationChannelIDs
	}

	tagIDs, monitorIDs, channelIDs, err = validateAlertRuleRelationships(c, user.ID, scope, tagIDs, monitorIDs, channelIDs)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	tx, err := database.Client.Tx(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start alert rule transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update alert rule",
		})
	}

	update := tx.AlertRule.UpdateOneID(id).
		ClearTags().
		ClearMonitors().
		ClearNotificationChannels()
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.Description != nil {
		if *req.Description == "" {
			update.ClearDescription()
		} else {
			update.SetDescription(*req.Description)
		}
	}
	if req.Event != nil {
		update.SetEvent(entalertrule.Event(*req.Event))
	}
	if req.Scope != nil {
		update.SetScope(scope)
	}
	if req.IsActive != nil {
		update.SetIsActive(*req.IsActive)
	}
	if req.ResendIntervalSeconds != nil {
		if *req.ResendIntervalSeconds == 0 {
			update.ClearResendIntervalSeconds()
		} else {
			update.SetResendIntervalSeconds(*req.ResendIntervalSeconds)
		}
	}
	if len(tagIDs) > 0 {
		update.AddTagIDs(tagIDs...)
	}
	if len(monitorIDs) > 0 {
		update.AddMonitorIDs(monitorIDs...)
	}
	if len(channelIDs) > 0 {
		update.AddNotificationChannelIDs(channelIDs...)
	}

	if _, err := update.Save(c); err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to update alert rule")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update alert rule",
		})
	}
	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit alert rule transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update alert rule",
		})
	}

	rule, err := userAlertRule(c, user.ID, id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload alert rule")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update alert rule",
		})
	}

	return c.Status(fiber.StatusOK).JSON(alertRuleResponse(rule))
}
