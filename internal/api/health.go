package api

import (
	"github.com/gofiber/fiber/v2"
)

type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// @Summary Health check
// @Description Check if the API is healthy
// @Tags Health
// @Success 200 {object} HealthCheckResponse "status and message"
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {

	// If everything is healthy, return a success response
	return c.JSON(HealthCheckResponse{
		Status:  "success",
		Message: "API is healthy",
	})
}
