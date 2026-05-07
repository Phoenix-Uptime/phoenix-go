package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type SettingsResponse struct {
	SMTPSettings *models.SMTPSettings `json:"smtp_settings,omitempty"`
	TelegramBot  *models.TelegramBot  `json:"telegram_bot,omitempty"`
}

// @Summary Get User Settings
// @Description Returns SMTP and Telegram bot settings for the authenticated user.
// @Tags Account
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} SettingsResponse "user settings"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/settings [get]
func GetAccountSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	smtpChannel, err := defaultNotificationChannel(c, user.ID, entnotificationchannel.TypeSMTP)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load SMTP settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	telegramChannel, err := defaultNotificationChannel(c, user.ID, entnotificationchannel.TypeTelegram)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Telegram bot settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	response := SettingsResponse{
		SMTPSettings: smtpSettingsFromNotificationChannel(smtpChannel),
		TelegramBot:  telegramBotFromNotificationChannel(telegramChannel),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

type UpdateSMTPSettingsRequest struct {
	SMTPServer  string `json:"smtp_server" validate:"required" swagger:"example=smtp.example.com"`
	SMTPPort    int    `json:"smtp_port" validate:"required" swagger:"example=587"`
	FromAddress string `json:"from_address" validate:"required,email" swagger:"example=noreply@example.com"`
	Username    string `json:"username" validate:"required" swagger:"example=user@example.com"`
	Password    string `json:"password" validate:"required" swagger:"example=supersecret"`
	UseTLS      bool   `json:"use_tls" swagger:"example=true"`
}

// @Summary Update SMTP Settings
// @Description Updates the SMTP settings for the authenticated user.
// @Tags Account
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body UpdateSMTPSettingsRequest true "SMTP settings"
// @Success 200 {object} SuccessResponse "settings updated"
// @Failure 400 {object} ErrorResponse "invalid input"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/settings/smtp [post]
func UpdateSMTPSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req UpdateSMTPSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	if err := upsertDefaultNotificationChannel(c, user.ID, entnotificationchannel.TypeSMTP, "SMTP", map[string]interface{}{
		"smtp_server":  req.SMTPServer,
		"smtp_port":    req.SMTPPort,
		"from_address": req.FromAddress,
		"username":     req.Username,
		"password":     req.Password,
		"use_tls":      req.UseTLS,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to update SMTP settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to update SMTP settings",
		})
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Status:  "success",
		Message: "SMTP settings updated",
	})
}

type UpdateTelegramBotRequest struct {
	BotToken string `json:"bot_token" validate:"required" swagger:"example=123456789:ABCdefGHIjklMNOpqrSTUvwxyz"`
}

// @Summary Update Telegram Bot Settings
// @Description Updates the Telegram bot settings for the authenticated user.
// @Tags Account
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Param data body UpdateTelegramBotRequest true "Telegram bot settings"
// @Success 200 {object} SuccessResponse "settings updated"
// @Failure 400 {object} ErrorResponse "invalid input"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/settings/telegram [post]
func UpdateTelegramBotSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req UpdateTelegramBotRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	if err := upsertDefaultNotificationChannel(c, user.ID, entnotificationchannel.TypeTelegram, "Telegram", map[string]interface{}{
		"bot_token": req.BotToken,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to update Telegram bot settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to update Telegram bot settings",
		})
	}

	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Status:  "success",
		Message: "Telegram bot settings updated",
	})
}

func defaultNotificationChannel(c fiber.Ctx, userID int, channelType entnotificationchannel.Type) (*ent.NotificationChannel, error) {
	channel, err := database.Client.NotificationChannel.Query().
		Where(
			entnotificationchannel.UserID(userID),
			entnotificationchannel.TypeEQ(channelType),
			entnotificationchannel.IsDefault(true),
		).
		First(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return channel, nil
}

func upsertDefaultNotificationChannel(c fiber.Ctx, userID int, channelType entnotificationchannel.Type, name string, config map[string]interface{}) error {
	channel, err := defaultNotificationChannel(c, userID, channelType)
	if err != nil {
		return err
	}
	if channel == nil {
		return database.Client.NotificationChannel.Create().
			SetUserID(userID).
			SetName(name).
			SetType(channelType).
			SetIsActive(true).
			SetIsDefault(true).
			SetConfig(config).
			Exec(c)
	}

	return database.Client.NotificationChannel.UpdateOneID(channel.ID).
		SetName(name).
		SetIsActive(true).
		SetIsDefault(true).
		SetConfig(config).
		Exec(c)
}

func smtpSettingsFromNotificationChannel(channel *ent.NotificationChannel) *models.SMTPSettings {
	if channel == nil || len(channel.Config) == 0 {
		return nil
	}

	config := channel.Config
	return &models.SMTPSettings{
		SMTPServer:  stringConfig(config, "smtp_server"),
		SMTPPort:    intConfig(config, "smtp_port"),
		FromAddress: stringConfig(config, "from_address"),
		Username:    stringConfig(config, "username"),
		Password:    stringConfig(config, "password"),
		UseTLS:      boolConfig(config, "use_tls"),
	}
}

func telegramBotFromNotificationChannel(channel *ent.NotificationChannel) *models.TelegramBot {
	if channel == nil || len(channel.Config) == 0 {
		return nil
	}

	botToken := stringConfig(channel.Config, "bot_token")
	if botToken == "" {
		return nil
	}
	return &models.TelegramBot{
		BotToken: botToken,
	}
}

func stringConfig(config map[string]interface{}, key string) string {
	value, ok := config[key].(string)
	if !ok {
		return ""
	}
	return value
}

func intConfig(config map[string]interface{}, key string) int {
	switch value := config[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func boolConfig(config map[string]interface{}, key string) bool {
	value, ok := config[key].(bool)
	if !ok {
		return false
	}
	return value
}
