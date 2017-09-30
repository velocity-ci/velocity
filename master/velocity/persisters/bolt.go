package persisters

import (
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

func newBoltDB(logger *log.Logger, dbPath string) *bolt.DB {

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Fatal(err)
	}

	return db
}

var boltDB *bolt.DB

func GetBoltDB() *bolt.DB {
	var dbPath string
	if os.Getenv("DB_PATH") != "" {
		dbPath = os.Getenv("DB_PATH")
	} else {
		dbPath = "cache.db"
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) || boltDB == nil {
		boltLogger := log.New(os.Stdout, "[bolt]", log.Lshortfile)
		boltDB = newBoltDB(boltLogger, dbPath)
	}

	return boltDB
}
