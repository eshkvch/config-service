package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig `validate:"required"`
	HTTP     HTTPConfig     `validate:"required"`
}

type DatabaseConfig struct {
	DSN string `validate:"required"`
}

type HTTPConfig struct {
	Port string `validate:"required"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dsn, err := databaseDSN()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Database: DatabaseConfig{
			DSN: dsn,
		},
		HTTP: HTTPConfig{
			Port: getEnvOrDefault("PORT", "8080"),
		},
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func databaseDSN() (string, error) {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn, nil
	}

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if user == "" || password == "" || dbName == "" {
		return "", fmt.Errorf("set DATABASE_URL or DB_USER, DB_PASSWORD, DB_NAME (and optionally DB_HOST, DB_PORT)")
	}

	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, dbName,
	), nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}
