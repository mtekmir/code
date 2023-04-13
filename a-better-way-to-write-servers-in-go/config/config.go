package config

import "time"

type Config struct {
	Port           string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
}

func Parse() (Config, error) {
	c := Config{
		// Default values
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1048576, // 1mb
		Port:           "8080",
	}

	// parse config values from env vars or flags
	// ...

	return c, nil
}
