package test

import (
	"database/sql"
	"os"
	"testing"

	"code.com/config"
)

// SetupTX sets up a database transaction to be used in tests.
// Once the tests are done it will rollback the transaction
func SetupTX(t *testing.T) *sql.Tx {
	t.Helper()

	conf := config.Parse()

	db, err := sql.Open("pgx", conf.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to initialize db. Err: %s", err.Error())
	}

	// create test schema
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS test")
	if err != nil {
		t.Fatalf("Error while creating the schema. Err: %s", err.Error())
	}

	// use schema
	_, err = db.Exec("SET search_path TO test")
	if err != nil {
		t.Fatalf("Error while switching to schema. Err: %s", err.Error())
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Unable to begin tx. %v", err)
	}

	t.Cleanup(func() {
		tx.Rollback()
		db.Close()
	})

	return tx
}

type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func CreateProductTables(t *testing.T, db execer) {
	t.Helper()
	statements := []string{
		`
		create table if not exists variations(
			id bigserial unique primary key,
			name varchar not null
		);
		`,
		`
		create table if not exists products(
			id bigserial unique primary key,
			barcode varchar not null,
			name varchar not null,
			price numeric(10,2),
			quantity int not null,
			variation_id bigint references variations(id)
		);
		`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("failed to create table. %v", err)
		}
	}
}
