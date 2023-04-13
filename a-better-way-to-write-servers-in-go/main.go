package main

import (
	"github/mtekmir/a-server/config"
	"github/mtekmir/a-server/server"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf, err := config.Parse()
	if err != nil {
		return err
	}
	s := server.New(conf)
	return s.Start()
}
