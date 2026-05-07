package notification_channels

import (
	"strings"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UpdateNotificationChannelRequest struct {
	Name             *string `json:"name,omitempty" validate:"omitempty,min=1"`
	IsActive         *bool   `json:"is_active,omitempty"`
	IsDefault        *bool   `json:"is_default,omitempty"`
	SMTPServer       *string `json:"smtp_server,omitempty"`
	SMTPPort         *int    `json:"smtp_port,omitempty" validate:"omitempty,min=1,max=65535"`
	SMTPFromAddress  *string `json:"smtp_from_address,omitempty" validate:"omitempty,email"`
	SMTPUsername     *string `json:"smtp_username,omitempty"`
	SMTPPassword     *string `json:"smtp_password,omitempty"`
	SMTPUseTLS       *bool   `json:"smtp_use_tls,omitempty"`
	TelegramBotToken *string `json:"telegram_bot_token,omitempty"`
	TelegramChatID   *string `json:"telegram_chat_id,omitempty"`
	WebhookURL       *string `json:"webhook_url,omitempty" validate:"omitempty,url"`
	WebhookMethod    *string `json:"webhook_method,omitempty"`
	NtfyServerURL    *string `json:"ntfy_server_url,omitempty" validate:"omitempty,url"`
	NtfyTopic        *string `json:"ntfy_topic,omitempty"`
	NtfyToken        *string `json:"ntfy_token,omitempty"`
}

// @Summary Update Notification Channel
// @Description Updates one notification channel owned by the authenticated user.
// @Tags Notification Channels
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param id path int true "Notification Channel ID"
// @Param data body UpdateNotificationChannelRequest true "notification channel payload"
// @Success 200 {object} NotificationChannelResponse "updated notification channel"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 404 {object} routes.ErrorResponse "notification channel not found"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /notification-channels/{id} [patch]
func UpdateNotificationChannel(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	id, err := notificationChannelID(c)
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

	var req UpdateNotificationChannelRequest
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
	if req.WebhookMethod != nil && *req.WebhookMethod != "" && !validWebhookMethod(*req.WebhookMethod) {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Webhook method is invalid",
		})
	}

	tx, err := database.Client.Tx(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start notification channel transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update notification channel",
		})
	}

	if req.IsDefault != nil && *req.IsDefault {
		if err := tx.NotificationChannel.Update().
			Where(
				entnotificationchannel.UserID(user.ID),
				entnotificationchannel.TypeEQ(channel.Type),
				entnotificationchannel.IDNEQ(id),
			).
			SetIsDefault(false).
			Exec(c); err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).Msg("Failed to clear default notification channels")
			return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Failed to update notification channel",
			})
		}
	}

	update := tx.NotificationChannel.UpdateOneID(id)
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.IsActive != nil {
		update.SetIsActive(*req.IsActive)
	}
	if req.IsDefault != nil {
		update.SetIsDefault(*req.IsDefault)
	}
	if req.SMTPServer != nil {
		if *req.SMTPServer == "" {
			update.ClearSMTPServer()
		} else {
			update.SetSMTPServer(*req.SMTPServer)
		}
	}
	if req.SMTPPort != nil {
		if *req.SMTPPort == 0 {
			update.ClearSMTPPort()
		} else {
			update.SetSMTPPort(*req.SMTPPort)
		}
	}
	if req.SMTPFromAddress != nil {
		if *req.SMTPFromAddress == "" {
			update.ClearSMTPFromAddress()
		} else {
			update.SetSMTPFromAddress(*req.SMTPFromAddress)
		}
	}
	if req.SMTPUsername != nil {
		if *req.SMTPUsername == "" {
			update.ClearSMTPUsername()
		} else {
			update.SetSMTPUsername(*req.SMTPUsername)
		}
	}
	if req.SMTPPassword != nil {
		if *req.SMTPPassword == "" {
			update.ClearSMTPPassword()
		} else {
			update.SetSMTPPassword(*req.SMTPPassword)
		}
	}
	if req.SMTPUseTLS != nil {
		update.SetSMTPUseTLS(*req.SMTPUseTLS)
	}
	if req.TelegramBotToken != nil {
		if *req.TelegramBotToken == "" {
			update.ClearTelegramBotToken()
		} else {
			update.SetTelegramBotToken(*req.TelegramBotToken)
		}
	}
	if req.TelegramChatID != nil {
		if *req.TelegramChatID == "" {
			update.ClearTelegramChatID()
		} else {
			update.SetTelegramChatID(*req.TelegramChatID)
		}
	}
	if req.WebhookURL != nil {
		if *req.WebhookURL == "" {
			update.ClearWebhookURL()
		} else {
			update.SetWebhookURL(*req.WebhookURL)
		}
	}
	if req.WebhookMethod != nil {
		if *req.WebhookMethod == "" {
			update.ClearWebhookMethod()
		} else {
			update.SetWebhookMethod(strings.ToUpper(*req.WebhookMethod))
		}
	}
	if req.NtfyServerURL != nil {
		if *req.NtfyServerURL == "" {
			update.ClearNtfyServerURL()
		} else {
			update.SetNtfyServerURL(*req.NtfyServerURL)
		}
	}
	if req.NtfyTopic != nil {
		if *req.NtfyTopic == "" {
			update.ClearNtfyTopic()
		} else {
			update.SetNtfyTopic(*req.NtfyTopic)
		}
	}
	if req.NtfyToken != nil {
		if *req.NtfyToken == "" {
			update.ClearNtfyToken()
		} else {
			update.SetNtfyToken(*req.NtfyToken)
		}
	}

	updated, err := update.Save(c)
	if err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to update notification channel")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update notification channel",
		})
	}

	response := notificationChannelResponse(updated)
	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit notification channel transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update notification channel",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
