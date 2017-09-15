package persisters

import (
	"log"
	"time"

	"github.com/boltdb/bolt"
)

func NewBoltDB(logger *log.Logger) *bolt.DB {

	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	return db
}
