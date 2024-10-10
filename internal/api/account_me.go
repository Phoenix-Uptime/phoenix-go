package api

import (
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/gofiber/fiber/v2"
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
// @Param api_key header string false "API Key for user authentication (Header)"
// @Param api_key query string false "API Key for user authentication (Query)"
// @Security ApiKeyHeader
// @Security ApiKeyQuery
// @Success 200 {object} UserResponse "user basic information"
// @Failure 401 {object} ErrorResponse "unauthorized - invalid or missing API key"
// @Router /account/me [get]
func GetAccountMe(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	response := UserResponse{
		Username: user.Username,
		Email:    user.Email,
		ApiKey:   user.ApiKey,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
