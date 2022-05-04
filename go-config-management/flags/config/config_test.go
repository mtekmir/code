package config_test

import (
	"code.com/config"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	os.Args[1] = "-db-host=hostname"
	os.Args[2] = "-db-port=1234"
	os.Args[3] = "-db-user=user"
	os.Args[4] = "-db-password=pass"

	c := config.Parse()

	if c.DBHost != "hostname" {
		t.Errorf("Expected dbHost to be 'hostname'. Got %s", c.DBHost)
	}

	if c.DBPort != "1234" {
		t.Errorf("Expected dbPort to be '1234'. Got %s", c.DBPort)
	}

	if c.DBUser != "user" {
		t.Errorf("Expected dbUser to be 'user'. Got %s", c.DBUser)
	}

	if c.DBPassword != "pass" {
		t.Errorf("Expected dbPassword to be 'pass'. Got %s", c.DBPassword)
	}
}
