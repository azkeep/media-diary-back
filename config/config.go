package config

import (
	"os"
)

type Config struct {
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	Port          string
	AllowedOrigin string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("PORT_DB", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "postgres"),
		DBName:        getEnv("DB_NAME", "matalogue"),
		Port:          getEnv("PORT_GO", "8080"),
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}
