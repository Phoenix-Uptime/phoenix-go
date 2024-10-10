package database

import (
	"fmt"

	"github.com/Phoenix-Uptime/phoenix-go/internal/config"
	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection based on the configured driver.
func InitDB() error {
	driver := config.GetDatabaseDriver()

	var err error
	switch driver {
	case "postgres":
		err = connectPostgres()
	case "sqlite":
		err = connectSQLite()
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().Msg("Database connection established successfully")
	return nil
}

func connectSQLite() error {
	path := config.GetSQLitePath()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to SQLite")
		return err
	}

	DB = db
	log.Info().Msgf("Connected to SQLite database at %s", path)
	return nil
}

func connectPostgres() error {
	host, port, user, password, dbname := config.GetPostgresConfig()

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return fmt.Errorf("incomplete PostgreSQL configuration")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to PostgreSQL")
		return err
	}

	DB = db
	log.Info().Msg("Connected to PostgreSQL database")
	return nil
}
