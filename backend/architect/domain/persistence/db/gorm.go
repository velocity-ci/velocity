package db

import (
	"log"
	"os"
	"sync"

	"github.com/jinzhu/gorm"

	// SQLite3
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB
var once sync.Once

func init() {
	once.Do(func() {
		var err error
		db, err = gorm.Open("sqlite3", os.Getenv("DB_PATH"))
		if err != nil {
			log.Fatal(err)
			panic("failed to connect database")
		}

		db.AutoMigrate(&user{}, &knownhost{}, &project{})
	})
}

func GetDialectName() string {
	return db.Dialect().GetName()
}

func Exec(q string) error {
	_, err := db.DB().Exec(q)
	return err
}

func Close() error {
	return db.Close()
}
