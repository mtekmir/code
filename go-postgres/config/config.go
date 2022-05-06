package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
}

func Parse() Config {
	c := Config{
		DatabaseURL: getEnvOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
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
