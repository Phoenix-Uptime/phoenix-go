package server

import (
	"time"

	accountroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/account"
	authroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/auth"
	healthroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/health"
	monitorroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/monitors"
	tagroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/tags"
	"github.com/Phoenix-Uptime/phoenix-go/internal/server/middleware"
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rs/zerolog/log"

	_ "github.com/Phoenix-Uptime/phoenix-go/docs"
)

func New() *fiber.App {
	app := fiber.New()

	// Set up CORS to allow all origins
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowCredentials: false,
	}))

	// Zerolog middleware for request logging
	app.Use(func(c fiber.Ctx) error {
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
	app.Get("/swagger/*", swaggo.HandlerDefault)

	// Register health check route
	app.Get("/health", healthroutes.HealthCheck)

	// Auth routes
	app.Post("/login", authroutes.Login)
	app.Post("/signup", authroutes.Signup)

	// Account routes
	account := app.Group("/account")
	account.Use(middleware.AuthMiddleware)
	account.Get("/me", accountroutes.GetAccountMe)
	account.Get("/settings", accountroutes.GetAccountSettings)
	account.Post("/reset-api-key", accountroutes.ResetAPIKey)
	account.Post("/change-password", accountroutes.ChangePassword)
	account.Post("/settings/telegram", accountroutes.UpdateTelegramBotSettings)
	account.Post("/settings/smtp", accountroutes.UpdateSMTPSettings)

	// Tag routes
	tags := app.Group("/tags")
	tags.Use(middleware.AuthMiddleware)
	tags.Get("", tagroutes.ListTags)
	tags.Post("", tagroutes.CreateTag)
	tags.Get("/:id", tagroutes.GetTag)
	tags.Patch("/:id", tagroutes.UpdateTag)
	tags.Delete("/:id", tagroutes.DeleteTag)

	// Monitor routes
	monitors := app.Group("/monitors")
	monitors.Use(middleware.AuthMiddleware)
	monitors.Get("", monitorroutes.ListMonitors)
	monitors.Post("", monitorroutes.CreateMonitor)
	monitors.Get("/:id", monitorroutes.GetMonitor)
	monitors.Patch("/:id", monitorroutes.UpdateMonitor)
	monitors.Delete("/:id", monitorroutes.DeleteMonitor)
	monitors.Post("/:id/pause", monitorroutes.PauseMonitor)
	monitors.Post("/:id/resume", monitorroutes.ResumeMonitor)
	monitors.Get("/:id/checks", monitorroutes.ListMonitorChecks)
	monitors.Get("/:id/stats", monitorroutes.ListMonitorStats)
	monitors.Get("/:id/tags", monitorroutes.ListMonitorTags)
	monitors.Put("/:id/tags", monitorroutes.ReplaceMonitorTags)

	return app
}
