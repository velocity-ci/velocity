package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/jmoiron/sqlx"
	"github.com/rakyll/statik/fs"

	// Static SQL migrations
	_ "github.com/velocity-ci/velocity/backend/pkg/grpc/architect/db/migrations"
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

	statikFS, err := fs.New()
	if err != nil {
		return nil, fmt.Errorf("unable to access statik data: %w", err)
	}

	var assetNames []string
	fs.Walk(statikFS, "/", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			assetNames = append(assetNames, info.Name())
		}

		return nil
	})

	s := bindata.Resource(assetNames, func(name string) ([]byte, error) {
		return fs.ReadFile(statikFS, filepath.Join("/", name))
	})

	d, err := bindata.WithInstance(s)
	m, err := migrate.NewWithInstance(
		"go-bindata", d,
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
