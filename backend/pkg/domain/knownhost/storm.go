package knownhost

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
	db.Init(&KnownHost{})
	return &stormDB{db}
}

func (db *stormDB) save(kH *KnownHost) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(kH); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *stormDB) delete(kH *KnownHost) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	tx.DeleteStruct(kH)

	return tx.Commit()
}

func (db *stormDB) exists(entry string) bool {
	query := db.Select(q.Eq("Entry", entry))
	var kH KnownHost
	if err := query.First(&kH); err != nil {
		return false
	}

	return true
}

func (db *stormDB) getAll(pQ *domain.PagingQuery) (r []*KnownHost, t int) {
	t = 0
	t, err := db.Count(&KnownHost{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	query := db.Select()
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	query.Find(&r)

	return r, t
}

func GetByUUID(db *storm.DB, uuid string) (*KnownHost, error) {
	var kH KnownHost
	if err := db.One("UUID", uuid, &kH); err != nil {
		return nil, err
	}
	return &kH, nil
}
