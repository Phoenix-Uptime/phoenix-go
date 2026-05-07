package server

import (
	"time"

	accountroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/account"
	alertruleroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/alert_rules"
	apikeyroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/api_keys"
	authroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/auth"
	healthroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/health"
	incidentroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/incidents"
	maintenancewindowroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/maintenance_windows"
	monitorroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/monitors"
	notificationchannelroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/notification_channels"
	statuspageroutes "github.com/Phoenix-Uptime/phoenix-go/internal/routes/status_pages"
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

	// API key routes
	apiKeys := app.Group("/api-keys")
	apiKeys.Use(middleware.AuthMiddleware)
	apiKeys.Get("", apikeyroutes.ListAPIKeys)
	apiKeys.Post("", apikeyroutes.CreateAPIKey)
	apiKeys.Get("/:id", apikeyroutes.GetAPIKey)
	apiKeys.Patch("/:id", apikeyroutes.UpdateAPIKey)
	apiKeys.Delete("/:id", apikeyroutes.DeleteAPIKey)

	// Tag routes
	tags := app.Group("/tags")
	tags.Use(middleware.AuthMiddleware)
	tags.Get("", tagroutes.ListTags)
	tags.Post("", tagroutes.CreateTag)
	tags.Get("/:id", tagroutes.GetTag)
	tags.Patch("/:id", tagroutes.UpdateTag)
	tags.Delete("/:id", tagroutes.DeleteTag)

	// Notification channel routes
	notificationChannels := app.Group("/notification-channels")
	notificationChannels.Use(middleware.AuthMiddleware)
	notificationChannels.Get("", notificationchannelroutes.ListNotificationChannels)
	notificationChannels.Post("", notificationchannelroutes.CreateNotificationChannel)
	notificationChannels.Get("/:id", notificationchannelroutes.GetNotificationChannel)
	notificationChannels.Patch("/:id", notificationchannelroutes.UpdateNotificationChannel)
	notificationChannels.Delete("/:id", notificationchannelroutes.DeleteNotificationChannel)

	// Alert rule routes
	alertRules := app.Group("/alert-rules")
	alertRules.Use(middleware.AuthMiddleware)
	alertRules.Get("", alertruleroutes.ListAlertRules)
	alertRules.Post("", alertruleroutes.CreateAlertRule)
	alertRules.Get("/:id", alertruleroutes.GetAlertRule)
	alertRules.Patch("/:id", alertruleroutes.UpdateAlertRule)
	alertRules.Delete("/:id", alertruleroutes.DeleteAlertRule)

	// Incident routes
	incidents := app.Group("/incidents")
	incidents.Use(middleware.AuthMiddleware)
	incidents.Get("", incidentroutes.ListIncidents)
	incidents.Post("", incidentroutes.CreateIncident)
	incidents.Get("/:id", incidentroutes.GetIncident)
	incidents.Patch("/:id", incidentroutes.UpdateIncident)
	incidents.Delete("/:id", incidentroutes.DeleteIncident)
	incidents.Post("/:id/acknowledge", incidentroutes.AcknowledgeIncident)
	incidents.Post("/:id/resolve", incidentroutes.ResolveIncident)

	// Maintenance window routes
	maintenanceWindows := app.Group("/maintenance-windows")
	maintenanceWindows.Use(middleware.AuthMiddleware)
	maintenanceWindows.Get("", maintenancewindowroutes.ListMaintenanceWindows)
	maintenanceWindows.Post("", maintenancewindowroutes.CreateMaintenanceWindow)
	maintenanceWindows.Get("/:id", maintenancewindowroutes.GetMaintenanceWindow)
	maintenanceWindows.Patch("/:id", maintenancewindowroutes.UpdateMaintenanceWindow)
	maintenanceWindows.Delete("/:id", maintenancewindowroutes.DeleteMaintenanceWindow)
	maintenanceWindows.Post("/:id/activate", maintenancewindowroutes.ActivateMaintenanceWindow)
	maintenanceWindows.Post("/:id/deactivate", maintenancewindowroutes.DeactivateMaintenanceWindow)

	// Status page routes
	statusPages := app.Group("/status-pages")
	statusPages.Use(middleware.AuthMiddleware)
	statusPages.Get("", statuspageroutes.ListStatusPages)
	statusPages.Post("", statuspageroutes.CreateStatusPage)
	statusPages.Get("/:id", statuspageroutes.GetStatusPage)
	statusPages.Patch("/:id", statuspageroutes.UpdateStatusPage)
	statusPages.Delete("/:id", statuspageroutes.DeleteStatusPage)
	statusPages.Get("/:id/monitors", statuspageroutes.ListStatusPageMonitors)
	statusPages.Put("/:id/monitors", statuspageroutes.ReplaceStatusPageMonitors)

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
