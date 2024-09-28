package main

import (
	"fmt"
	"os"

	"github.com/Phoenix-Uptime/phoenix-go/internal/database"
	"github.com/Phoenix-Uptime/phoenix-go/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := database.InitDB(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}

	app := server.New()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8484"
	}

	log.Info().Msgf("Starting server on :%s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
