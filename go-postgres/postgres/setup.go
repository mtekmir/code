package postgres

import (
	"database/sql"

	// Postgres driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Setup sets up the db and returns a func to close it.
func Setup(dbURL string) (*sql.DB, func(), error) {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return nil, nil, err
	}

	closeDB := func() { db.Close() }

	return db, closeDB, nil
}
