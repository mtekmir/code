package main

import (
	"log"

	"code.com/config"
	"code.com/postgres"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf := config.Parse()

	_, dbClose, err := postgres.Setup(conf.DatabaseURL)
	if err != nil {
		return err
	}
	defer dbClose()

	return nil
}
