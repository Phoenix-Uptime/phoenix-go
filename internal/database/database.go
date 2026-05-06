package database

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/Phoenix-Uptime/phoenix-go/ent"
	"github.com/Phoenix-Uptime/phoenix-go/internal/config"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

var Client *ent.Client

// InitDB initializes the Ent client and creates missing schema resources.
func InitDB() error {
	cfg := config.GetDatabase()

	var (
		client *ent.Client
		err    error
	)
	switch cfg.Driver {
	case "postgres":
		client, err = connectPostgres(cfg.Postgres)
	case "sqlite", "sqlite3":
		client, err = connectSQLite(cfg.SQLite)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := client.Schema.Create(context.Background()); err != nil {
		_ = client.Close()
		return fmt.Errorf("failed to migrate schemas: %w", err)
	}

	Client = client
	log.Info().Msg("Database connection established successfully")
	return nil
}

func Close() error {
	if Client == nil {
		return nil
	}
	return Client.Close()
}

func connectSQLite(cfg *config.SQLiteConfig) (*ent.Client, error) {
	path := cfg.Path
	db, err := stdsql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}

	log.Info().Msgf("Connected to SQLite database at %s", path)
	return ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db))), nil
}

func connectPostgres(cfg *config.PostgresConfig) (*ent.Client, error) {
	if cfg.Host == "" || cfg.Port == 0 || cfg.DBName == "" {
		return nil, fmt.Errorf("incomplete PostgreSQL configuration")
	}

	dsn := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.DBName)
	if cfg.User != "" {
		dsn += fmt.Sprintf(" user=%s password=%s", cfg.User, cfg.Password)
	}

	client, err := ent.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Connected to PostgreSQL database")
	return client, nil
}
