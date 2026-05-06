package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
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

	user, err := database.Client.User.Get(c, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load user settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	response := SettingsResponse{
		SMTPSettings: smtpSettingsFromUser(user),
		TelegramBot:  telegramBotFromUser(user),
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

	if err := database.Client.User.UpdateOneID(user.ID).
		SetSMTPSMTPServer(req.SMTPServer).
		SetSMTPSMTPPort(req.SMTPPort).
		SetSMTPFromAddress(req.FromAddress).
		SetSMTPUsername(req.Username).
		SetSMTPPassword(req.Password).
		SetSMTPUseTLS(req.UseTLS).
		Exec(c); err != nil {
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

	if err := database.Client.User.UpdateOneID(user.ID).
		SetTelegramBotToken(req.BotToken).
		Exec(c); err != nil {
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

func smtpSettingsFromUser(user *ent.User) *models.SMTPSettings {
	if user.SMTPSMTPServer == "" &&
		user.SMTPSMTPPort == 0 &&
		user.SMTPFromAddress == "" &&
		user.SMTPUsername == "" &&
		user.SMTPPassword == "" &&
		!user.SMTPUseTLS {
		return nil
	}

	return &models.SMTPSettings{
		SMTPServer:  user.SMTPSMTPServer,
		SMTPPort:    user.SMTPSMTPPort,
		FromAddress: user.SMTPFromAddress,
		Username:    user.SMTPUsername,
		Password:    user.SMTPPassword,
		UseTLS:      user.SMTPUseTLS,
	}
}

func telegramBotFromUser(user *ent.User) *models.TelegramBot {
	if user.TelegramBotToken == "" {
		return nil
	}

	return &models.TelegramBot{
		BotToken: user.TelegramBotToken,
	}
}
