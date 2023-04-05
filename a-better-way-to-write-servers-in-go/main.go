package main

import (
	"github/mtekmir/a-server/server"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	s := server.New()
	return s.Start()
}
