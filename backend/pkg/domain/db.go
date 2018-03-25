package domain

import (
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
)

func NewStormDB(path string) *storm.DB {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	db, err := storm.Open(path)
	if err != nil {
		logrus.Fatal(err)
	}

	return db
}

type PagingQuery struct {
	Limit int `json:"amount" query:"amount"`
	Page  int `json:"page" query:"page"`
}

func NewPagingQuery() *PagingQuery {
	return &PagingQuery{
		Limit: 10,
		Page:  1,
	}
}
