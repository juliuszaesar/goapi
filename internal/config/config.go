package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Typesense TypesenseConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
	SSLMode  string
}

// TypesenseConfig holds Typesense configuration
type TypesenseConfig struct {
	Host   string
	Port   string
	APIKey string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "goapi"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "goapi"),
			Port:     getEnv("DB_PORT", "5432"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Typesense: TypesenseConfig{
			Host:   getEnv("TYPESENSE_HOST", "localhost"),
			Port:   getEnv("TYPESENSE_PORT", "8108"),
			APIKey: getEnv("TYPESENSE_API_KEY", "xyz"),
		},
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return "host=" + c.Host + " user=" + c.User + " password=" + c.Password + 
		   " dbname=" + c.DBName + " port=" + c.Port + " sslmode=" + c.SSLMode
}

// GetTypesenseURL returns the Typesense URL
func (c *TypesenseConfig) GetTypesenseURL() string {
	return "http://" + c.Host + ":" + c.Port
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
