package domain

import (
	"log"

	"github.com/asdine/storm"
	"github.com/jinzhu/gorm"
	// SQLite3
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func NewGORMDB(path string) *gorm.DB {
	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
		panic("failed to connect database")
	}

	// db.LogMode(true)

	return db
}

func NewStormDB(path string) *storm.DB {
	db, err := storm.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

type PagingQuery struct {
	Limit int
	Page  int
}
