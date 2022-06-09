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

func CreateProductTables(t *testing.T, db *sql.DB) {
	t.Helper()
	statements := []string{
		`create table if not exists brands(
			id bigserial unique primary key,
			name varchar not null
		);`,
		`create table if not exists variations(
			id bigserial unique primary key,
			name varchar not null
		);`,
		`create table if not exists products(
			id bigserial unique primary key,
			created_at timestamptz default now(),
			name varchar not null,
			price int not null,
			brand_id bigint references brands(id)
		);`,
		`create table if not exists product_variations(
			id bigserial unique primary key,
			quantity int not null,
			product_id bigint not null references products(id),
			variation_id bigint not null references variations(id)
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("failed to create table. %v", err)
		}
	}
}
