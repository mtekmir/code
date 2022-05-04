package main

import (
	"fmt"

	"code.com/config"
)

func main() {
	c := config.Parse()

	fmt.Println(c.DBHost)
}
