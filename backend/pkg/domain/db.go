package domain

import (
	"log"

	"github.com/asdine/storm"
)

func NewStormDB(path string) *storm.DB {
	db, err := storm.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

type PagingQuery struct {
	Limit int `json:"limit"`
	Page  int `json:"page"`
}

func NewPagingQuery() *PagingQuery {
	return &PagingQuery{
		Limit: 10,
		Page:  1,
	}
}
