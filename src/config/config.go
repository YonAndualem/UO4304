// Package config loads application configuration from environment variables.
// Using environment variables (rather than hard-coded values or config files)
// follows the 12-Factor App methodology, which makes the application easy to
// configure across different environments (local, staging, production) without
// changing code. Docker and docker-compose inject these via env_file or environment.
package config

import (
	"fmt"
	"os"
)

// Config holds all runtime configuration for the application.
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load reads configuration from environment variables, falling back to safe
// development defaults when a variable is not set.
func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "trade_license"),
		DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}
}

// DSN builds the PostgreSQL Data Source Name string expected by GORM's postgres driver.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
