package config

import (
	"os"
)

type Config struct {
	DBHost     string // Host of database server
	DBPort     string // ...
	DBUser     string
	DBPassword string
	// ...
}

func Parse() Config {
	c := Config{
		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvOrDefault("DB_PORT", "5432"),
		DBUser:     getEnvOrDefault("DB_USER", "postgres"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", "postgres"),
	}
	return c
}

func getEnvOrDefault(key string, defaultVal string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	return v
}
