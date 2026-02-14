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

	cfg := &Config{
		Database: DatabaseConfig{
			DSN: os.Getenv("DATABASE_URL"),
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
