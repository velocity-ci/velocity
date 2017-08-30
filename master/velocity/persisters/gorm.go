package persisters

import (
	"fmt"
	"log"
	"os"

	"github.com/VJftw/velocity/master/velocity/utils"
	"github.com/jinzhu/gorm"
	// databases
	_ "github.com/jinzhu/gorm/dialects/mysql"
	// _ "github.com/jinzhu/gorm/dialects/postgres"
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
)

// NewGORMDB - Initialises a connection to a GORM storage
func NewGORMDB(logger *log.Logger, models ...interface{}) *gorm.DB {

	if !utils.WaitForService(fmt.Sprintf("%s:%s", os.Getenv("DATABASE_HOST"), os.Getenv("DATABASE_PORT")), logger) {
		panic("Could not find database")
	}

	var db *gorm.DB
	var err error

	if os.Getenv("DATABASE_TYPE") == "MYSQL" {
		db, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
			os.Getenv("DATABASE_USERNAME"),
			os.Getenv("DATABASE_PASSWORD"),
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_NAME"),
		))
	} else if os.Getenv("DATABASE_TYPE") == "POSTGRES" {
		db, err = gorm.Open("postgres", fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_DBNAME"),
			os.Getenv("POSTGRES_PASSWORD"),
		))
	} else if os.Getenv("DATABASE_TYPE") == "SQLITE" {
		db, err = gorm.Open("sqlite3", "/tmp/gorm.db")
	}

	if db == nil {
		panic("Database not defined")
	}

	if err != nil {
		fmt.Println(err)
		panic("failed to connect database")
	}

	db.AutoMigrate(models...)

	if os.Getenv("GORM_DEBUG") == "true" {
		db.LogMode(true)
	}

	return db
}
