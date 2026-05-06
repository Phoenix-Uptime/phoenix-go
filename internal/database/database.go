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
	driver := config.GetDatabaseDriver()

	var (
		client *ent.Client
		err    error
	)
	switch driver {
	case "postgres":
		client, err = connectPostgres()
	case "sqlite":
		client, err = connectSQLite()
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
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

func connectSQLite() (*ent.Client, error) {
	path := config.GetSQLitePath()
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

func connectPostgres() (*ent.Client, error) {
	host, port, user, password, dbname := config.GetPostgresConfig()
	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("incomplete PostgreSQL configuration")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	client, err := ent.Open(dialect.Postgres, dsn)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Connected to PostgreSQL database")
	return client, nil
}
