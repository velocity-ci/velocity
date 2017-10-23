package main

import (
	"log"
	"time"

	"github.com/boltdb/bolt"
)

func NewBoltDB(logger *log.Logger, dbPath string) *bolt.DB {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Fatal(err)
	}

	return db
}
