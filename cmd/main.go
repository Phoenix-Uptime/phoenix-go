package main

import (
	"fmt"
	"os"

	"github.com/Phoenix-Uptime/phoenix-go/internal/config"
	"github.com/Phoenix-Uptime/phoenix-go/internal/models"
	"github.com/Phoenix-Uptime/phoenix-go/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title PhoenixUptime API
// @version 1.0
// @description PhoenixUptime Backend API

// @host 127.0.0.1:3031
// @BasePath /

// @securityDefinitions.apikey ApiKeyHeader
// @type        apiKey
// @name        x-api-key
// @in          header

// @securityDefinitions.apikey ApiKeyQuery
// @type        apiKey
// @name        api_key
// @in          query
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize the configuration
	if err := config.InitConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize configuration")
	}

	// Initialize the database connection
	if err := models.InitDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Initialize the server
	app := server.New()

	// Get host and port from the configuration
	host, port := config.GetServerConfig()
	address := fmt.Sprintf("%s:%s", host, port)

	log.Info().Msgf("Starting server on %s", address)
	if err := app.Listen(address); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
