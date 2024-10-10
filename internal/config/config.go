package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var (
	k        = koanf.New(".")
	validate = validator.New()
	config   = &ConfigStructure{}
)

type ConfigStructure struct {
	Server   ServerConfig   `validate:"required"`
	Database DatabaseConfig `validate:"required"`
}

type ServerConfig struct {
	Host string `validate:"required,hostname_rfc1123"`
	Port string `validate:"required,numeric"`
}

type DatabaseConfig struct {
	Driver   string         `validate:"required,oneof=sqlite postgres"`
	SQLite   SQLiteConfig   `validate:"required_if=Driver sqlite"`
	Postgres PostgresConfig `validate:"required_if=Driver postgres"`
}

type SQLiteConfig struct {
	Path string `validate:"required_if=Database.Driver sqlite"`
}

type PostgresConfig struct {
	Host     string `validate:"required_if=Database.Driver postgres"`
	Port     string `validate:"required_if=Database.Driver postgres,numeric"`
	User     string `validate:"required_if=Database.Driver postgres"`
	Password string `validate:"required_if=Database.Driver postgres"`
	DBName   string `validate:"required_if=Database.Driver postgres"`
}

// InitConfig loads configuration from a file with optional environment overrides
func InitConfig() error {
	if err := k.Load(file.Provider("phoenix.yaml"), yaml.Parser()); err != nil {
		return errors.New("missing required configuration file: phoenix.yaml")
	}

	// Load environment variable overrides with "PHOENIX_" prefix
	err := k.Load(env.Provider("PHOENIX_", ".", func(key string) string {
		// Convert environment variables like PHOENIX_DATABASE_SQLITE_PATH to database.sqlite.path
		return strings.Replace(strings.ToLower(strings.TrimPrefix(key, "PHOENIX_")), "_", ".", -1)
	}), nil)
	if err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Unmarshal to apply all settings into the config struct
	if err := k.Unmarshal("", config); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate the final configuration
	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}
