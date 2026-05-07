package alert_rules

import (
	"fmt"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entalertrule "github.com/Phoenix-Uptime/phoenix-go/ent/alertrule"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type CreateAlertRuleRequest struct {
	Name                   string  `json:"name" validate:"required,min=1"`
	Description            *string `json:"description,omitempty"`
	Event                  string  `json:"event" validate:"required,oneof=down recovered degraded certificate_expiring"`
	Scope                  string  `json:"scope" validate:"omitempty,oneof=all tags monitors"`
	IsActive               *bool   `json:"is_active,omitempty"`
	ResendIntervalSeconds  *int    `json:"resend_interval_seconds,omitempty" validate:"omitempty,min=1"`
	TagIDs                 []int   `json:"tag_ids,omitempty"`
	MonitorIDs             []int   `json:"monitor_ids,omitempty"`
	NotificationChannelIDs []int   `json:"notification_channel_ids" validate:"required,min=1"`
}

// @Summary Create Alert Rule
// @Description Creates an alert rule owned by the authenticated user.
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateAlertRuleRequest true "alert rule payload"
// @Success 201 {object} AlertRuleResponse "created alert rule"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /alert-rules [post]
func CreateAlertRule(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateAlertRuleRequest
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

	scope := entalertrule.ScopeAll
	if req.Scope != "" {
		scope = entalertrule.Scope(req.Scope)
	}
	tagIDs, monitorIDs, channelIDs, err := validateAlertRuleRelationships(c, user.ID, scope, req.TagIDs, req.MonitorIDs, req.NotificationChannelIDs)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	create := database.Client.AlertRule.Create().
		SetUserID(user.ID).
		SetName(req.Name).
		SetEvent(entalertrule.Event(req.Event)).
		SetScope(scope).
		AddNotificationChannelIDs(channelIDs...)
	if req.Description != nil {
		create.SetDescription(*req.Description)
	}
	if req.IsActive != nil {
		create.SetIsActive(*req.IsActive)
	}
	if req.ResendIntervalSeconds != nil {
		create.SetResendIntervalSeconds(*req.ResendIntervalSeconds)
	}
	if len(tagIDs) > 0 {
		create.AddTagIDs(tagIDs...)
	}
	if len(monitorIDs) > 0 {
		create.AddMonitorIDs(monitorIDs...)
	}

	rule, err := create.Save(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create alert rule")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create alert rule",
		})
	}

	rule, err = userAlertRule(c, user.ID, rule.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reload alert rule")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create alert rule",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(alertRuleResponse(rule))
}

func validateAlertRuleRelationships(c fiber.Ctx, userID int, scope entalertrule.Scope, tagIDs []int, monitorIDs []int, channelIDs []int) ([]int, []int, []int, error) {
	tagIDs, ok := normalizeIDs(tagIDs)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid tag ids")
	}
	monitorIDs, ok = normalizeIDs(monitorIDs)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid monitor ids")
	}
	channelIDs, ok = normalizeIDs(channelIDs)
	if !ok || len(channelIDs) == 0 {
		return nil, nil, nil, fmt.Errorf("invalid notification channel ids")
	}

	switch scope {
	case entalertrule.ScopeTags:
		if len(tagIDs) == 0 {
			return nil, nil, nil, fmt.Errorf("tag ids are required for tag scoped alert rules")
		}
		monitorIDs = nil
	case entalertrule.ScopeMonitors:
		if len(monitorIDs) == 0 {
			return nil, nil, nil, fmt.Errorf("monitor ids are required for monitor scoped alert rules")
		}
		tagIDs = nil
	case entalertrule.ScopeAll:
		tagIDs = nil
		monitorIDs = nil
	default:
		return nil, nil, nil, fmt.Errorf("invalid alert rule scope")
	}

	ok, err := userOwnsTagIDs(c, userID, tagIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate alert rule tags")
		return nil, nil, nil, fmt.Errorf("failed to validate tag ids")
	}
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid tag ids")
	}
	ok, err = userOwnsMonitorIDs(c, userID, monitorIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate alert rule monitors")
		return nil, nil, nil, fmt.Errorf("failed to validate monitor ids")
	}
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid monitor ids")
	}
	ok, err = userOwnsNotificationChannelIDs(c, userID, channelIDs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate alert rule notification channels")
		return nil, nil, nil, fmt.Errorf("failed to validate notification channel ids")
	}
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid notification channel ids")
	}

	return tagIDs, monitorIDs, channelIDs, nil
}
