package db

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func NewDB() (*DB, error) {
	db, err := sqlx.Open("postgres", "host=localhost port=5432 user=velocity password=velocity dbname=velocity sslmode=disable")
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/configs/sql/migrations", cwd),
		"postgres", driver)
	if err != nil {
		return nil, err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	return &DB{db}, nil
}
