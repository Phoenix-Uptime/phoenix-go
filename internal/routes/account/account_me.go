package account

import (
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/gofiber/fiber/v3"
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

	response := UserResponse{
		Username: user.Username,
		Email:    user.Email,
		ApiKey:   user.APIKey,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
