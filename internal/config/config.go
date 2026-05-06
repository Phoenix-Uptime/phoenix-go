package config

import (
	"fmt"
	"os"
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
	Port uint   `validate:"required,port"`
}

type DatabaseConfig struct {
	Driver   string          `validate:"required,oneof=sqlite sqlite3 postgres"`
	SQLite   *SQLiteConfig   `validate:"required_unless=Driver postgres"`
	Postgres *PostgresConfig `validate:"required_if=Driver postgres"`
}

type SQLiteConfig struct {
	Path string `validate:"required,filepath"`
}

type PostgresConfig struct {
	Host     string `validate:"required,hostname|ip"`
	Port     uint   `validate:"required,port"`
	User     string `validate:"omitempty,printascii"`
	Password string `validate:"required_with=User,omitempty,printascii"`
	DBName   string `validate:"required,printascii"`
}

// InitConfig loads configuration from a file with optional environment overrides
func InitConfig() error {
	k.Set("server.host", "127.0.0.1")
	k.Set("server.port", 3031)
	k.Set("database.driver", "sqlite")
	k.Set("database.sqlite.path", "phoenix.db")
	k.Set("database.postgres.host", "localhost")
	k.Set("database.postgres.port", 5432)
	k.Set("database.postgres.dbname", "phoenix")

	if _, err := os.Stat("phoenix.yaml"); err == nil {
		if err := k.Load(file.Provider("phoenix.yaml"), yaml.Parser()); err != nil {
			return fmt.Errorf("failed to load configuration file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect configuration file: %w", err)
	}

	// Load environment variable overrides with "PHOENIX_" prefix
	err := k.Load(env.Provider("PHOENIX_", ".", func(key string) string {
		key = strings.TrimPrefix(key, "PHOENIX_")
		key = strings.ToLower(key)
		key = strings.ReplaceAll(key, "_", ".")
		return key
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

func Get() *ConfigStructure {
	return config
}

func GetServer() ServerConfig {
	return config.Server
}

func GetDatabase() DatabaseConfig {
	return config.Database
}
