package knownhost

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
)

type gormKnownHost struct {
	UUID              string `gorm:"primary_key"`
	Entry             string
	Hosts             []byte
	Comment           string
	SHA256Fingerprint string
	MD5Fingerprint    string
}

func (gormKnownHost) TableName() string {
	return "knownhosts"
}

func (gK *gormKnownHost) toKnownHost() *KnownHost {
	hosts := []string{}
	err := json.Unmarshal(gK.Hosts, &hosts)
	if err != nil {
		logrus.Error(err)
	}
	return &KnownHost{
		UUID:              gK.UUID,
		Entry:             gK.Entry,
		Hosts:             hosts,
		Comment:           gK.Comment,
		SHA256Fingerprint: gK.SHA256Fingerprint,
		MD5Fingerprint:    gK.MD5Fingerprint,
	}
}

func (k *KnownHost) toGormKnownHost() *gormKnownHost {
	jsonHosts, err := json.Marshal(k.Hosts)
	if err != nil {
		logrus.Error(err)
	}

	return &gormKnownHost{
		UUID:              k.UUID,
		Entry:             k.Entry,
		Hosts:             jsonHosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

type db struct {
	db *gorm.DB
}

func newDB(gorm *gorm.DB) *db {
	return &db{
		db: gorm,
	}
}

func (db *db) delete(k *KnownHost) error {
	tx := db.db.Begin()

	gK := k.toGormKnownHost()

	if err := tx.Delete(gK).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (db *db) save(k *KnownHost) error {
	tx := db.db.Begin()

	gK := k.toGormKnownHost()

	tx.
		Where(gormKnownHost{Entry: k.Entry}).
		Assign(&gK).
		FirstOrCreate(&gK)

	return tx.Commit().Error
}

func (db *db) getAll(q *domain.PagingQuery) (r []*KnownHost, t int) {
	t = 0

	gKs := []gormKnownHost{}
	db.db.Find(&gKs).Count(&t)

	db.db.
		Limit(q.Limit).
		Offset((q.Page - 1) * q.Limit).
		Find(&gKs)

	for _, gK := range gKs {
		r = append(r, gK.toKnownHost())
	}

	return r, t
}

func (db *db) exists(entry string) bool {
	if db.db.Where("entry = ?", entry).First(&gormKnownHost{}).RecordNotFound() {
		return false
	}
	return true
}
