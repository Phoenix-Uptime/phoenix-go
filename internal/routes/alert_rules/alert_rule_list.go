package alert_rules

import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entalertrule "github.com/Phoenix-Uptime/phoenix-go/ent/alertrule"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type AlertRuleListResponse struct {
	AlertRules []AlertRuleResponse `json:"alert_rules"`
}

// @Summary List Alert Rules
// @Description Returns alert rules owned by the authenticated user.
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param event query string false "down, recovered, degraded, or certificate_expiring"
// @Param scope query string false "all, tags, or monitors"
// @Success 200 {object} AlertRuleListResponse "alert rule list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /alert-rules [get]
func ListAlertRules(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.AlertRule.Query().
		Where(entalertrule.UserID(user.ID)).
		WithTags().
		WithMonitors().
		WithNotificationChannels().
		Order(entalertrule.ByCreatedAt(entsql.OrderDesc()))

	if event := c.Query("event"); event != "" {
		ruleEvent := entalertrule.Event(event)
		if err := entalertrule.EventValidator(ruleEvent); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid alert rule event",
			})
		}
		query.Where(entalertrule.EventEQ(ruleEvent))
	}
	if scope := c.Query("scope"); scope != "" {
		ruleScope := entalertrule.Scope(scope)
		if err := entalertrule.ScopeValidator(ruleScope); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Invalid alert rule scope",
			})
		}
		query.Where(entalertrule.ScopeEQ(ruleScope))
	}

	rules, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list alert rules")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to list alert rules",
		})
	}

	response := AlertRuleListResponse{
		AlertRules: make([]AlertRuleResponse, 0, len(rules)),
	}
	for _, rule := range rules {
		response.AlertRules = append(response.AlertRules, alertRuleResponse(rule))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
