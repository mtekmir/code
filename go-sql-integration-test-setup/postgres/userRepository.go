package postgres

import (
	"context"
	"database/sql"
)

type DB interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type UserRepo struct {
	db DB
}

func NewUserRepo(db DB) UserRepo {
	return UserRepo{db: db}
}

func (r UserRepo) UserExists(ctx context.Context, ID int) (bool, error) {
	var exists bool

	sql := `SELECT EXISTS(SELECT true FROM users WHERE id=$1)`
	if err := r.db.QueryRowContext(ctx, sql, ID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r UserRepo) GetAll(ctx context.Context) ([]string, error) {
	var nn []string

	rows, err := r.db.QueryContext(ctx, `SELECT name from users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		nn = append(nn, n)
	}

	return nn, nil
}
