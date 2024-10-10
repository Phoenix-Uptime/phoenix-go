package server

import (
	"time"

	"github.com/Phoenix-Uptime/phoenix-go/internal/api"
	"github.com/Phoenix-Uptime/phoenix-go/internal/server/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/rs/zerolog/log"

	_ "github.com/Phoenix-Uptime/phoenix-go/docs"
)

func New() *fiber.App {
	app := fiber.New()

	// Set up CORS to allow all origins
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: false,
	}))

	// Zerolog middleware for request logging
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()

		// Proceed to the next middleware or handler
		err := c.Next()

		// Log request details
		logEvent := log.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP())

		if err != nil {
			logEvent.Err(err)
		}

		logEvent.Msg("Request handled")
		return err
	})

	// Register Swagger route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Register health check route
	app.Get("/health", api.HealthCheck)

	// Auth routes
	app.Post("/login", api.Login)
	app.Post("/signup", api.Signup)

	// Account routes
	account := app.Group("/account")
	account.Use(middleware.AuthMiddleware)
	account.Get("/me", api.GetAccountMe)
	account.Get("/settings", api.GetAccountSettings)
	account.Post("/reset-api-key", api.ResetAPIKey)
	account.Post("/change-password", api.ChangePassword)
	account.Post("/settings/settings/telegram", api.UpdateTelegramBotSettings)
	account.Post("/settings/settings/smtp", api.UpdateSMTPSettings)

	return app
}
