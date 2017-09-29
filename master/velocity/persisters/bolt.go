package persisters

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

func newBoltDB(logger *log.Logger) *bolt.DB {

	db, err := bolt.Open("cache.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Fatal(err)
	}

	return db
}

var boltDB *bolt.DB
var once sync.Once

func GetBoltDB() *bolt.DB {
	once.Do(func() {
		boltLogger := log.New(os.Stdout, "[bolt]", log.Lshortfile)
		boltDB = newBoltDB(boltLogger)
	})

	return boltDB
}
