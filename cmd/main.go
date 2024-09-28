package main

import (
	"fmt"
	"os"

	"github.com/Phoenix-Uptime/phoenix-go/internal/api"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	_ "github.com/Phoenix-Uptime/phoenix-go/docs"
)

// @title PhoenixUptime API
// @version 1.0
// @description PhoenixUptime Backend API

// @host 127.0.0.1:8484
// @BasePath /

// @securityDefinitions.apikey ApiKeyHeader
// @type        apiKey
// @name        x-api-key
// @in          header

// @securityDefinitions.apikey SessionIDHeader
// @type        apiKey
// @name        x-session-id
// @in          header

// @securityDefinitions.apikey ApiKeyQuery
// @type        apiKey
// @name        api_key
// @in          query

// @securityDefinitions.apikey SessionIDQuery
// @type        apiKey
// @name        sid
// @in          query
func main() {
	// Set zerolog time format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Use ConsoleWriter for human-readable logs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize Fiber app
	app := fiber.New()

	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open("phoenix.db"), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to SQLite database")
	}

	// Run database migrations (for example purposes, no specific models)
	if err := db.AutoMigrate(); err != nil {
		log.Fatal().Err(err).Msg("Failed to auto-migrate")
	}

	// Swagger docs route
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check endpoint
	app.Get("/health", api.HealthCheck)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8484" // Default port
	}

	// Start server
	log.Info().Msgf("Starting server on :%s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
