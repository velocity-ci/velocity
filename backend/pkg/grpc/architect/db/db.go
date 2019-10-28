package db

import pg "github.com/go-pg/pg/v9"

type DB struct {
	*pg.DB
}

func NewDB() *DB {
	db := pg.Connect(&pg.Options{
		Addr:     "localhost:26257",
		User:     "admin",
		Password: "admin",
		Database: "postgres",
	})

	return &DB{db}
}
