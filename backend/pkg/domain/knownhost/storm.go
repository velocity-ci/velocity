package knownhost

import (
	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type stormKnownHost struct {
	ID                string `storm:"id"`
	Entry             string
	Hosts             []string
	Comment           string
	SHA256Fingerprint string
	MD5Fingerprint    string
}

func (s *stormKnownHost) ToKnownHost() *KnownHost {
	return &KnownHost{
		ID:                s.ID,
		Entry:             s.Entry,
		Hosts:             s.Hosts,
		Comment:           s.Comment,
		SHA256Fingerprint: s.SHA256Fingerprint,
		MD5Fingerprint:    s.MD5Fingerprint,
	}
}

func (k *KnownHost) toStormKnownHost() *stormKnownHost {
	return &stormKnownHost{
		ID:                k.ID,
		Entry:             k.Entry,
		Hosts:             k.Hosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

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

	if err := tx.Save(kH.toStormKnownHost()); err != nil {
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

	tx.DeleteStruct(kH.toStormKnownHost())

	return tx.Commit()
}

func (db *stormDB) exists(entry string) bool {
	query := db.Select(q.Eq("Entry", entry))
	var kH stormKnownHost
	if err := query.First(&kH); err != nil {
		return false
	}

	return true
}

func (db *stormDB) getAll(pQ *domain.PagingQuery) (r []*KnownHost, t int) {
	t = 0
	t, err := db.Count(&stormKnownHost{})
	if err != nil {
		logrus.Error(err)
		return r, t
	}

	query := db.Select()
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var stormKnownHosts []*stormKnownHost
	query.Find(&stormKnownHosts)

	for _, k := range stormKnownHosts {
		r = append(r, k.ToKnownHost())
	}

	return r, t
}

func GetByID(db *storm.DB, id string) (*KnownHost, error) {
	var kH stormKnownHost
	if err := db.One("ID", id, &kH); err != nil {
		return nil, err
	}
	return kH.ToKnownHost(), nil
}
