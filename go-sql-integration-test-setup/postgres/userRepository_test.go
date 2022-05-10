package postgres_test

import (
	"context"
	"testing"

	"code.com/postgres"
	"code.com/test"
)

func TestUserExists(t *testing.T) {
	db := test.SetupTX(t)
	createUsersTable(t, db)
	repo := postgres.NewUserRepo(db)

	if _, err := db.Exec(`INSERT INTO users(name) VALUES ('mert')`); err != nil {
		t.Fatalf("failed to insert user. %v", err)
	}

	exists, err := repo.UserExists(context.TODO(), 1)
	if err != nil {
		t.Fatalf("UserExists() = %v", err)
	}

	if !exists {
		t.Error("UserExists() to be true")
	}
}

func TestGetAll(t *testing.T) {
	db := test.SetupTX(t)
	createUsersTable(t, db)
	repo := postgres.NewUserRepo(db)

	_, err := db.Exec(`
		INSERT INTO users(name) VALUES ('mert'), ('m'), ('t')
	`)
	if err != nil {
		t.Fatalf("failed to insert user. %v", err)
	}

	uu, err := repo.GetAll(context.TODO())
	if err != nil {
		t.Fatalf("GetAll() = %v", err)
	}

	if uu[1] != "mert" && uu[2] != "m" && uu[2] != "t" {
		t.Errorf("mismatch. %v", uu)
	}
}

func createUsersTable(t *testing.T, db postgres.DB) {
	t.Helper()

	_, err := db.Exec(`
	create table if not exists users(
		id bigserial unique primary key,
		name varchar unique not null
	)`)
	if err != nil {
		t.Fatalf("failed to create users table. %v", err)
	}
}
