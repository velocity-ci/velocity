package project

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type stormDB struct {
	*storm.DB
}

func newStormDB(db *storm.DB) *stormDB {
	db.Init(&Project{})
	return &stormDB{db}
}

func (db *stormDB) save(p *Project) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(p); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *stormDB) delete(p *Project) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	tx.DeleteStruct(p)

	return tx.Commit()
}

func (db *stormDB) getBySlug(slug string) (*Project, error) {
	query := db.Select(q.Eq("Slug", slug))
	var p Project
	if err := query.First(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (db *stormDB) getByName(name string) (*Project, error) {
	query := db.Select(q.Eq("Name", name))
	var p Project
	if err := query.First(&p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (db *stormDB) getAll(pQ *domain.PagingQuery) (r []*Project, t int) {
	t = 0
	t, err := db.Count(&Project{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	query := db.Select()
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

	return r, t
}

func GetByUUID(db *storm.DB, uuid string) (*Project, error) {
	var p Project
	if err := db.One("UUID", uuid, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
