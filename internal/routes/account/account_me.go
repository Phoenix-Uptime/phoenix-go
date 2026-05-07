package account

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/ent"
	entapikey "github.com/Phoenix-Uptime/phoenix-go/ent/apikey"
	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

type UserResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	ApiKey   string `json:"api_key"`
}

// @Summary Get User Information
// @Description Returns basic information about the authenticated user.
// @Tags Account
// @Accept json
// @Produce json
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} UserResponse "user basic information"
// @Failure 401 {object} routes.ErrorResponse "unauthorized - invalid or missing API key"
// @Router /account/me [get]
func GetAccountMe(c fiber.Ctx) error {
	user := c.Locals("user").(*ent.User)
	apiKey := ""
	key, err := database.Client.APIKey.Query().
		Where(
			entapikey.UserID(user.ID),
			entapikey.IsActive(true),
			entapikey.Or(
				entapikey.ExpiresAtIsNil(),
				entapikey.ExpiresAtGT(time.Now()),
			),
		).
		First(c)
	if err == nil {
		apiKey = key.Key
	} else if !ent.IsNotFound(err) {
		log.Error().Err(err).Msg("Failed to load user API key")
	}

	response := UserResponse{
		Username: user.Username,
		Email:    user.Email,
		ApiKey:   apiKey,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
