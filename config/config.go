package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Http
	Port          string
	AllowedOrigin string

	// Application and storage
	DefaultPageLimit   int
	MaxPageLimit       int
	ExportDir          string
	ExportDirPerm      os.FileMode
	ExportPrefixFormat string
	ExportRangeFormat  string
}

func Load() *Config {
	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("PORT_DB", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "matalogue"),
		Port:               getEnv("PORT_GO", "8080"),
		AllowedOrigin:      getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		DefaultPageLimit:   getEnvAsInt("DEFAULT_PAGE_LIMIT", 50),
		MaxPageLimit:       getEnvAsInt("MAX_PAGE_LIMIT", 100),
		ExportDir:          getEnv("EXPORT_DIR", "./export"),
		ExportDirPerm:      0755,
		ExportPrefixFormat: getEnv("EXPORT_PREFIX_FORMAT", "20060102150405"),
		ExportRangeFormat:  getEnv("EXPORT_RANGE_FORMAT", "20060102"),
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if valStr, exists := os.LookupEnv(key); exists {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return fallback
}
