package test

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

// SetupDB sets up a database connection to be used in tests.
// It creates a new schema with the t.Name().
// Once the test is complete, it will drop the created schema and close the db connection.
func SetupDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	u := os.Getenv("DATABASE_URL")
	if u != "" {
		dbURL = u
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		t.Fatalf("db initialization failed. err: %v", err)
	}

	schemaName := strings.ToLower(t.Name())

	t.Cleanup(func() {
		_, err := db.Exec("DROP SCHEMA " + schemaName + " CASCADE")
		if err != nil {
			t.Fatalf("db cleanup failed. err: %v", err)
		}
		db.Close()
	})

	// create test schema
	_, err = db.Exec("CREATE SCHEMA " + schemaName)
	if err != nil {
		t.Fatalf("schema creation failed. err: %v", err)
	}

	// use schema
	_, err = db.Exec("SET search_path TO " + schemaName)
	if err != nil {
		t.Fatalf("error while switching to schema. err: %v", err)
	}

	return db
}

func CreateProductTable(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS products(
			id bigserial unique primary key,
			created_at timestamptz default now(),
			name varchar not null
		)
	`)
	if err != nil {
		t.Fatalf("failed to create product table. error: %v", err)
	}
}
