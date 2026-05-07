package account

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entnotificationchannel "github.com/Phoenix-Uptime/phoenix-go/ent/notificationchannel"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/Phoenix-Uptime/phoenix-go/internal/routes"
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
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /account/settings [get]
func GetAccountSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	smtpChannel, err := defaultNotificationChannel(c, user.ID, entnotificationchannel.TypeSMTP)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load SMTP settings")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	telegramChannel, err := defaultNotificationChannel(c, user.ID, entnotificationchannel.TypeTelegram)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load Telegram bot settings")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
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
// @Success 200 {object} routes.SuccessResponse "settings updated"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /account/settings/smtp [post]
func UpdateSMTPSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req UpdateSMTPSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	if err := upsertDefaultSMTPChannel(c, user.ID, req); err != nil {
		log.Error().Err(err).Msg("Failed to update SMTP settings")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update SMTP settings",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
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
// @Success 200 {object} routes.SuccessResponse "settings updated"
// @Failure 400 {object} routes.ErrorResponse "invalid input"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} routes.ErrorResponse "internal server error"
// @Router /account/settings/telegram [post]
func UpdateTelegramBotSettings(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)

	var req UpdateTelegramBotRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid request payload",
		})
	}

	validate := validator.New()
	if err := validate.Struct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Invalid input: " + err.Error(),
		})
	}

	if err := upsertDefaultTelegramChannel(c, user.ID, req); err != nil {
		log.Error().Err(err).Msg("Failed to update Telegram bot settings")
		return c.Status(fiber.StatusInternalServerError).JSON(routes.ErrorResponse{
			Status:  "error",
			Message: "Failed to update Telegram bot settings",
		})
	}

	return c.Status(fiber.StatusOK).JSON(routes.SuccessResponse{
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

func upsertDefaultSMTPChannel(c fiber.Ctx, userID int, req UpdateSMTPSettingsRequest) error {
	channel, err := defaultNotificationChannel(c, userID, entnotificationchannel.TypeSMTP)
	if err != nil {
		return err
	}
	if channel == nil {
		return database.Client.NotificationChannel.Create().
			SetUserID(userID).
			SetName("SMTP").
			SetType(entnotificationchannel.TypeSMTP).
			SetIsActive(true).
			SetIsDefault(true).
			SetSMTPServer(req.SMTPServer).
			SetSMTPPort(req.SMTPPort).
			SetSMTPFromAddress(req.FromAddress).
			SetSMTPUsername(req.Username).
			SetSMTPPassword(req.Password).
			SetSMTPUseTLS(req.UseTLS).
			Exec(c)
	}

	return database.Client.NotificationChannel.UpdateOneID(channel.ID).
		SetName("SMTP").
		SetIsActive(true).
		SetIsDefault(true).
		SetSMTPServer(req.SMTPServer).
		SetSMTPPort(req.SMTPPort).
		SetSMTPFromAddress(req.FromAddress).
		SetSMTPUsername(req.Username).
		SetSMTPPassword(req.Password).
		SetSMTPUseTLS(req.UseTLS).
		Exec(c)
}

func upsertDefaultTelegramChannel(c fiber.Ctx, userID int, req UpdateTelegramBotRequest) error {
	channel, err := defaultNotificationChannel(c, userID, entnotificationchannel.TypeTelegram)
	if err != nil {
		return err
	}
	if channel == nil {
		return database.Client.NotificationChannel.Create().
			SetUserID(userID).
			SetName("Telegram").
			SetType(entnotificationchannel.TypeTelegram).
			SetIsActive(true).
			SetIsDefault(true).
			SetTelegramBotToken(req.BotToken).
			Exec(c)
	}

	return database.Client.NotificationChannel.UpdateOneID(channel.ID).
		SetName("Telegram").
		SetIsActive(true).
		SetIsDefault(true).
		SetTelegramBotToken(req.BotToken).
		Exec(c)
}

func smtpSettingsFromNotificationChannel(channel *ent.NotificationChannel) *models.SMTPSettings {
	if channel == nil ||
		channel.SMTPServer == nil &&
			channel.SMTPPort == nil &&
			channel.SMTPFromAddress == nil &&
			channel.SMTPUsername == nil &&
			channel.SMTPPassword == nil &&
			channel.SMTPUseTLS == nil {
		return nil
	}

	return &models.SMTPSettings{
		SMTPServer:  stringValue(channel.SMTPServer),
		SMTPPort:    intValue(channel.SMTPPort),
		FromAddress: stringValue(channel.SMTPFromAddress),
		Username:    stringValue(channel.SMTPUsername),
		Password:    stringValue(channel.SMTPPassword),
		UseTLS:      boolValue(channel.SMTPUseTLS),
	}
}

func telegramBotFromNotificationChannel(channel *ent.NotificationChannel) *models.TelegramBot {
	if channel == nil || channel.TelegramBotToken == nil {
		return nil
	}

	botToken := stringValue(channel.TelegramBotToken)
	if botToken == "" {
		return nil
	}
	return &models.TelegramBot{
		BotToken: botToken,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func boolValue(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}
