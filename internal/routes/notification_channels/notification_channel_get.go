package notification_channels

import (
	"strconv"
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type NotificationChannelResponse struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	Name                string    `json:"name"`
	Type                string    `json:"type"`
	IsActive            bool      `json:"is_active"`
	IsDefault           bool      `json:"is_default"`
	SMTPServer          *string   `json:"smtp_server,omitempty"`
	SMTPPort            *int      `json:"smtp_port,omitempty"`
	SMTPFromAddress     *string   `json:"smtp_from_address,omitempty"`
	SMTPUsername        *string   `json:"smtp_username,omitempty"`
	SMTPUseTLS          *bool     `json:"smtp_use_tls,omitempty"`
	HasSMTPPassword     bool      `json:"has_smtp_password"`
	HasTelegramBotToken bool      `json:"has_telegram_bot_token"`
	TelegramChatID      *string   `json:"telegram_chat_id,omitempty"`
	HasWebhookURL       bool      `json:"has_webhook_url"`
	WebhookMethod       *string   `json:"webhook_method,omitempty"`
	NtfyServerURL       *string   `json:"ntfy_server_url,omitempty"`
	NtfyTopic           *string   `json:"ntfy_topic,omitempty"`
	HasNtfyToken        bool      `json:"has_ntfy_token"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// @Summary Get Notification Channel
// @Description Returns one notification channel owned by the authenticated user.
// @Tags Notification Channels
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Notification Channel ID"
// @Success 200 {object} NotificationChannelResponse "notification channel"
// @Failure 400 {object} routes.ErrorResponse "invalid notification channel id"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "notification channel not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /notification-channels/{id} [get]
func GetNotificationChannel(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid notification channel id",
		})
	}

	channel, err := userNotificationChannel(c, user.ID, id)
	if err != nil {
		return notificationChannelLookupError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(notificationChannelResponse(channel))
}

func userNotificationChannel(c fiber.Ctx, userID int, id int) (*ent.NotificationChannel, error) {
	return database.Client.NotificationChannel.Query().
		Where(
			entnotificationchannel.ID(id),
			entnotificationchannel.UserID(userID),
		).
		Only(c)
}

func notificationChannelLookupError(c fiber.Ctx, err error) error {
	if ent.IsNotFound(err) {
		return c.Status(fiber.StatusNotFound).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Notification channel not found",
		})
	}
	log.Error().Err(err).Msg("Failed to load notification channel")
	return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
		Status:  "error",
		Message: "Failed to load notification channel",
	})
}

func notificationChannelResponse(channel *ent.NotificationChannel) NotificationChannelResponse {
	return NotificationChannelResponse{
		ID:                  channel.ID,
		UserID:              channel.UserID,
		Name:                channel.Name,
		Type:                string(channel.Type),
		IsActive:            channel.IsActive,
		IsDefault:           channel.IsDefault,
		SMTPServer:          channel.SMTPServer,
		SMTPPort:            channel.SMTPPort,
		SMTPFromAddress:     channel.SMTPFromAddress,
		SMTPUsername:        channel.SMTPUsername,
		SMTPUseTLS:          channel.SMTPUseTLS,
		HasSMTPPassword:     channel.SMTPPassword != nil && *channel.SMTPPassword != "",
		HasTelegramBotToken: channel.TelegramBotToken != nil && *channel.TelegramBotToken != "",
		TelegramChatID:      channel.TelegramChatID,
		HasWebhookURL:       channel.WebhookURL != nil && *channel.WebhookURL != "",
		WebhookMethod:       channel.WebhookMethod,
		NtfyServerURL:       channel.NtfyServerURL,
		NtfyTopic:           channel.NtfyTopic,
		HasNtfyToken:        channel.NtfyToken != nil && *channel.NtfyToken != "",
		CreatedAt:           channel.CreatedAt,
		UpdatedAt:           channel.UpdatedAt,
	}
}
