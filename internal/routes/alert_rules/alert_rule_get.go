package alert_rules

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entalertrule "github.com/Phoenix-Uptime/phoenix-go/ent/alertrule"
	entmonitor "github.com/Phoenix-Uptime/phoenix-go/ent/monitor"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	enttag "github.com/Phoenix-Uptime/phoenix-go/ent/tag"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type AlertRuleResponse struct {
	ID                     int       `json:"id"`
	UserID                 int       `json:"user_id"`
	Name                   string    `json:"name"`
	Description            *string   `json:"description,omitempty"`
	Event                  string    `json:"event"`
	Scope                  string    `json:"scope"`
	IsActive               bool      `json:"is_active"`
	ResendIntervalSeconds  *int      `json:"resend_interval_seconds,omitempty"`
	TagIDs                 []int     `json:"tag_ids"`
	MonitorIDs             []int     `json:"monitor_ids"`
	NotificationChannelIDs []int     `json:"notification_channel_ids"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// @Summary Get Alert Rule
// @Description Returns one alert rule owned by the authenticated user.
// @Tags Alert Rules
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Alert Rule ID"
// @Success 200 {object} AlertRuleResponse "alert rule"
// @Failure 400 {object} routes.ErrorResponse "invalid alert rule id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "alert rule not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /alert-rules/{id} [get]
func GetAlertRule(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := alertRuleID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid alert rule id",
		})
	}

	rule, err := userAlertRule(c, user.ID, id)
	if err != nil {
		return alertRuleLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(alertRuleResponse(rule))
}

func alertRuleID(c fiber.Ctx) (int, error) {
	return strconv.Atoi(c.Params("id"))
}

func userAlertRule(c fiber.Ctx, userID int, id int) (*ent.AlertRule, error) {
	return database.Client.AlertRule.Query().
		Where(
			entalertrule.ID(id),
			entalertrule.UserID(userID),
		).
		WithTags().
		WithMonitors().
		WithNotificationChannels().
		Only(c)
}

func alertRuleLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Alert rule not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load alert rule")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load alert rule",
	})
}

func alertRuleResponse(rule *ent.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
		ID:                     rule.ID,
		UserID:                 rule.UserID,
		Name:                   rule.Name,
		Description:            rule.Description,
		Event:                  string(rule.Event),
		Scope:                  string(rule.Scope),
		IsActive:               rule.IsActive,
		ResendIntervalSeconds:  rule.ResendIntervalSeconds,
		TagIDs:                 tagIDs(rule.Edges.Tags),
		MonitorIDs:             monitorIDs(rule.Edges.Monitors),
		NotificationChannelIDs: notificationChannelIDs(rule.Edges.NotificationChannels),
		CreatedAt:              rule.CreatedAt,
		UpdatedAt:              rule.UpdatedAt,
	}
}

func tagIDs(tags []*ent.Tag) []int {
	ids := make([]int, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}

func monitorIDs(monitors []*ent.Monitor) []int {
	ids := make([]int, 0, len(monitors))
	for _, monitor := range monitors {
		ids = append(ids, monitor.ID)
	}
	return ids
}

func notificationChannelIDs(channels []*ent.NotificationChannel) []int {
	ids := make([]int, 0, len(channels))
	for _, channel := range channels {
		ids = append(ids, channel.ID)
	}
	return ids
}

func normalizeIDs(ids []int) ([]int, bool) {
	seen := make(map[int]struct{}, len(ids))
	normalized := make([]int, 0, len(ids))
	for _, id := range ids {
		if id < 1 {
			return nil, false
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, true
}

func userOwnsTagIDs(c fiber.Ctx, userID int, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	count, err := database.Client.Tag.Query().
		Where(
			enttag.UserID(userID),
			enttag.IDIn(ids...),
		).
		Count(c)
	return count == len(ids), err
}

func userOwnsMonitorIDs(c fiber.Ctx, userID int, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	count, err := database.Client.Monitor.Query().
		Where(
			entmonitor.UserID(userID),
			entmonitor.IDIn(ids...),
		).
		Count(c)
	return count == len(ids), err
}

func userOwnsNotificationChannelIDs(c fiber.Ctx, userID int, ids []int) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	count, err := database.Client.NotificationChannel.Query().
		Where(
			entnotificationchannel.UserID(userID),
			entnotificationchannel.IDIn(ids...),
		).
		Count(c)
	return count == len(ids), err
}
