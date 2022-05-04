package config

import (
	"flag"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	// ...
}

func Parse() Config {
	dbHost := flag.String("db-host", "localhost", "database host.")
	dbPort := flag.String("db-port", "5432", "database port.")
	dbUser := flag.String("db-user", "postgres", "database user.")
	dbPass := flag.String("db-password", "postgres", "database password.")

	flag.Parse()

	c := Config{
		DBHost:     *dbHost,
		DBPort:     *dbPort,
		DBUser:     *dbUser,
		DBPassword: *dbPass,
	}

	return c
}
