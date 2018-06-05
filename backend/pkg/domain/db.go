package domain

import (
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	"github.com/golang/glog"
)

func NewStormDB(path string) *storm.DB {
	dir := filepath.Dir(path)
	os.MkdirAll(dir, os.ModePerm)
	db, err := storm.Open(path)
	if err != nil {
		glog.Fatal(err)
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
