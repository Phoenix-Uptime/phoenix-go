package notification_channels

import (
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type NotificationChannelListResponse struct {
	NotificationChannels []NotificationChannelResponse `json:"notification_channels"`
}

// @Summary List Notification Channels
// @Description Returns notification channels owned by the authenticated user.
// @Tags Notification Channels
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param type query string false "smtp, telegram, webhook, or ntfy"
// @Success 200 {object} NotificationChannelListResponse "notification channel list"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /notification-channels [get]
func ListNotificationChannels(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	query := database.Client.NotificationChannel.Query().
		Where(entnotificationchannel.UserID(user.ID)).
		Order(
			entnotificationchannel.ByType(entsql.OrderAsc()),
			entnotificationchannel.ByName(entsql.OrderAsc()),
		)

	if channelType := c.Query("type"); channelType != "" {
		typ := entnotificationchannel.Type(channelType)
		if err := entnotificationchannel.TypeValidator(typ); err != nil {
			return badRequest(c, "Invalid notification channel type")
		}
		query.Where(entnotificationchannel.TypeEQ(typ))
	}

	channels, err := query.All(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list notification channels")
		return serverError(c, "Failed to list notification channels")
	}

	response := NotificationChannelListResponse{
		NotificationChannels: make([]NotificationChannelResponse, 0, len(channels)),
	}
	for _, channel := range channels {
		response.NotificationChannels = append(response.NotificationChannels, notificationChannelResponse(channel))
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
