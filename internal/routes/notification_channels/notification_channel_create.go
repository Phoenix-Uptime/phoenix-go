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

type CreateNotificationChannelRequest struct {
	Name             string  `json:"name" validate:"required,min=1"`
	Type             string  `json:"type" validate:"required,oneof=smtp telegram webhook ntfy"`
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

// @Summary Create Notification Channel
// @Description Creates a notification channel owned by the authenticated user.
// @Tags Notification Channels
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body CreateNotificationChannelRequest true "notification channel payload"
// @Success 201 {object} NotificationChannelResponse "created notification channel"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /notification-channels [post]
func CreateNotificationChannel(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req CreateNotificationChannelRequest
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

	channelType := entnotificationchannel.Type(req.Type)
	if err := validateCreateNotificationChannel(req, channelType); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
	}

	tx, err := database.Client.Tx(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start notification channel transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create notification channel",
		})
	}

	isDefault := req.IsDefault != nil && *req.IsDefault
	if isDefault {
		if err := tx.NotificationChannel.Update().
			Where(
				entnotificationchannel.UserID(user.ID),
				entnotificationchannel.TypeEQ(channelType),
			).
			SetIsDefault(false).
			Exec(c); err != nil {
			_ = tx.Rollback()
			log.Error().Err(err).Msg("Failed to clear default notification channels")
			return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
				Status:  "error",
				Message: "Failed to create notification channel",
			})
		}
	}

	create := tx.NotificationChannel.Create().
		SetUserID(user.ID).
		SetName(req.Name).
		SetType(channelType).
		SetIsDefault(isDefault)
	if req.IsActive != nil {
		create.SetIsActive(*req.IsActive)
	}
	if req.SMTPServer != nil {
		create.SetSMTPServer(*req.SMTPServer)
	}
	if req.SMTPPort != nil {
		create.SetSMTPPort(*req.SMTPPort)
	}
	if req.SMTPFromAddress != nil {
		create.SetSMTPFromAddress(*req.SMTPFromAddress)
	}
	if req.SMTPUsername != nil {
		create.SetSMTPUsername(*req.SMTPUsername)
	}
	if req.SMTPPassword != nil {
		create.SetSMTPPassword(*req.SMTPPassword)
	}
	if req.SMTPUseTLS != nil {
		create.SetSMTPUseTLS(*req.SMTPUseTLS)
	}
	if req.TelegramBotToken != nil {
		create.SetTelegramBotToken(*req.TelegramBotToken)
	}
	if req.TelegramChatID != nil {
		create.SetTelegramChatID(*req.TelegramChatID)
	}
	if req.WebhookURL != nil {
		create.SetWebhookURL(*req.WebhookURL)
	}
	if req.WebhookMethod != nil {
		create.SetWebhookMethod(strings.ToUpper(*req.WebhookMethod))
	}
	if req.NtfyServerURL != nil {
		create.SetNtfyServerURL(*req.NtfyServerURL)
	}
	if req.NtfyTopic != nil {
		create.SetNtfyTopic(*req.NtfyTopic)
	}
	if req.NtfyToken != nil {
		create.SetNtfyToken(*req.NtfyToken)
	}

	channel, err := create.Save(c)
	if err != nil {
		_ = tx.Rollback()
		log.Error().Err(err).Msg("Failed to create notification channel")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create notification channel",
		})
	}

	response := notificationChannelResponse(channel)
	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("Failed to commit notification channel transaction")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to create notification channel",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func validateCreateNotificationChannel(req CreateNotificationChannelRequest, channelType entnotificationchannel.Type) error {
	if err := entnotificationchannel.TypeValidator(channelType); err != nil {
		return err
	}

	switch channelType {
	case entnotificationchannel.TypeSMTP:
		if missing(req.SMTPServer) || req.SMTPPort == nil || missing(req.SMTPFromAddress) {
			return fiber.NewError(fiber.StatusBadRequest, "SMTP server, port, and from address are required")
		}
	case entnotificationchannel.TypeTelegram:
		if missing(req.TelegramBotToken) || missing(req.TelegramChatID) {
			return fiber.NewError(fiber.StatusBadRequest, "Telegram bot token and chat id are required")
		}
	case entnotificationchannel.TypeWebhook:
		if missing(req.WebhookURL) {
			return fiber.NewError(fiber.StatusBadRequest, "Webhook URL is required")
		}
		if req.WebhookMethod != nil && !validWebhookMethod(*req.WebhookMethod) {
			return fiber.NewError(fiber.StatusBadRequest, "Webhook method is invalid")
		}
	case entnotificationchannel.TypeNtfy:
		if missing(req.NtfyServerURL) || missing(req.NtfyTopic) {
			return fiber.NewError(fiber.StatusBadRequest, "ntfy server URL and topic are required")
		}
	}

	return nil
}

func missing(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func validWebhookMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}
