package config_test

import (
	"os"
	"testing"

	"code.com/config"
)

func TestParse(t *testing.T) {
	t.Cleanup(func() {
		os.Clearenv()
	})

	os.Setenv("DB_HOST", "hostname")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_USER", "user")
	os.Setenv("DB_PASSWORD", "pass")

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
