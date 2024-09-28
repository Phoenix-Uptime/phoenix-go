package database

import (
	"fmt"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		driver = "sqlite"
	}

	var err error
	switch driver {
	case "postgres":
		err = connectPostgres()
	default:
		err = connectSQLite()
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().Msg("Database connection established successfully")
	return nil
}

func connectSQLite() error {
	dsn := "phoenix.db"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to SQLite")
		return err
	}

	DB = db
	log.Info().Msg("Connected to SQLite database")
	return nil
}

func connectPostgres() error {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")

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
