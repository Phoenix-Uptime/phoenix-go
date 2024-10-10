package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

type SMTPSettingsResponse struct {
	SMTPServer string `json:"smtp_server,omitempty"`
	SMTPPort   int    `json:"smtp_port,omitempty"`
	Username   string `json:"username,omitempty"`
}

type UserResponse struct {
	Username     string                `json:"username"`
	Email        string                `json:"email"`
	ApiKey       string                `json:"api_key"`
	SMTPSettings *SMTPSettingsResponse `json:"smtp_settings,omitempty"`
	TelegramBot  *models.TelegramBot   `json:"telegram_bot,omitempty"`
}

// @Summary Get User Information
// @Description Returns information about the authenticated user, including their SMTP and Telegram bot settings.
// @Tags Account
// @Accept json
// @Produce json
// @Param api_key header string false "API Key for user authentication (Header)"
// @Param api_key query string false "API Key for user authentication (Query)"
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} UserResponse "user information with settings"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/me [get]
func GetAccountInfo(c *fiber.Ctx) error {
	// Retrieve the authenticated user from the context
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		log.Error().Msg("User not found in context")
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
			Status:  "error",
			Message: "Unauthorized",
		})
	}

	if err := models.DB.Preload("SMTPSettings").Preload("TelegramBot").First(&user, user.ID).Error; err != nil {
		log.Error().Err(err).Msg("Failed to load user settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	smtpSettings := &SMTPSettingsResponse{}
	if user.SMTPSettings != nil {
		smtpSettings.SMTPServer = user.SMTPSettings.SMTPServer
		smtpSettings.SMTPPort = user.SMTPSettings.SMTPPort
		smtpSettings.Username = user.SMTPSettings.Username
	} else {
		smtpSettings = nil
	}

	response := UserResponse{
		Username:     user.Username,
		Email:        user.Email,
		ApiKey:       user.ApiKey,
		SMTPSettings: smtpSettings,
		TelegramBot:  user.TelegramBot,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
