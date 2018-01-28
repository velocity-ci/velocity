package main

import (
	"log"

	"github.com/jinzhu/gorm"
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
