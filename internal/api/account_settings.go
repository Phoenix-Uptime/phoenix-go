package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/gofiber/fiber/v2"
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
// @Param api_key header string false "API Key for user authentication (Header)"
// @Param api_key query string false "API Key for user authentication (Query)"
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} SettingsResponse "user settings"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /account/settings [get]
func GetAccountSettings(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	// Ensure settings are loaded
	if err := models.DB.Preload("SMTPSettings").Preload("TelegramBot").First(&user, user.ID).Error; err != nil {
		log.Error().Err(err).Msg("Failed to load user settings")
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Status:  "error",
			Message: "Failed to load user settings",
		})
	}

	response := SettingsResponse{
		SMTPSettings: user.SMTPSettings,
		TelegramBot:  user.TelegramBot,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
