package domain

import (
	"log"
	"os"
	"path/filepath"

	"github.com/asdine/storm"
)

func NewStormDB(path string) *storm.DB {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	db, err := storm.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

type PagingQuery struct {
	Limit int `json:"limit" query:"limit"`
	Page  int `json:"page" query:"page"`
}

func NewPagingQuery() *PagingQuery {
	return &PagingQuery{
		Limit: 10,
		Page:  1,
	}
}
