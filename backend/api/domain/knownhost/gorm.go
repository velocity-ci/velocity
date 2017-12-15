package knownhost

import (
	"fmt"
	"log"
	"os"

	"github.com/docker/go/canonical/json"

	"github.com/jinzhu/gorm"
)

type gormKnownHost struct {
	ID                string `gorm:"primary_key"`
	Entry             string
	Hosts             []byte
	Comment           string
	SHA256Fingerprint string
	MD5Fingerprint    string
}

func (gormKnownHost) TableName() string {
	return "knownhosts"
}

func knownHostFromGormKnownHost(gK gormKnownHost) KnownHost {
	hosts := []string{}
	err := json.Unmarshal(gK.Hosts, &hosts)
	if err != nil {
		log.Println("Could not unmarshal knownhost hosts")
		log.Fatal(err)
	}
	return KnownHost{
		ID:                gK.ID,
		Entry:             gK.Entry,
		Hosts:             hosts,
		Comment:           gK.Comment,
		SHA256Fingerprint: gK.SHA256Fingerprint,
		MD5Fingerprint:    gK.MD5Fingerprint,
	}
}

func gormKnownHostFromKnownHost(k KnownHost) gormKnownHost {
	jsonHosts, err := json.Marshal(k.Hosts)
	if err != nil {
		log.Println("Could not marshal knownhost hosts")
		log.Fatal(err)
	}

	return gormKnownHost{
		ID:                k.ID,
		Entry:             k.Entry,
		Hosts:             jsonHosts,
		Comment:           k.Comment,
		SHA256Fingerprint: k.SHA256Fingerprint,
		MD5Fingerprint:    k.MD5Fingerprint,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(gormKnownHost{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:knownhost]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(k KnownHost) KnownHost {
	tx := r.gorm.Begin()

	gK := gormKnownHostFromKnownHost(k)

	err := tx.Where(&gormKnownHost{
		ID: k.ID,
	}).First(&gormKnownHost{}).Error
	if err != nil {
		err = tx.Create(&gK).Error
	} else {
		tx.Save(&gK)
	}

	tx.Commit()
	r.logger.Printf("saved knownhost %s", k.ID)

	return knownHostFromGormKnownHost(gK)
}

func (r *gormRepository) Delete(k KnownHost) {
	tx := r.gorm.Begin()

	gK := gormKnownHostFromKnownHost(k)

	if err := tx.Delete(gK).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByID(id string) (KnownHost, error) {
	gK := gormKnownHost{}

	if r.gorm.
		Where(&gormKnownHost{
			ID: id,
		}).
		First(&gK).RecordNotFound() {
		r.logger.Printf("could not find knownhost %s", id)
		return KnownHost{}, fmt.Errorf("could not find knownhost %s", id)
	}

	r.logger.Printf("got knownhost %s", id)
	return knownHostFromGormKnownHost(gK), nil
}

func (r *gormRepository) GetAll(q KnownHostQuery) ([]KnownHost, uint64) {

	gKs := []gormKnownHost{}
	var count uint64

	db := r.gorm

	db.
		Find(&gKs).
		Count(&count)

	db.
		Limit(int(q.Amount)).
		Offset(int(q.Page - 1)).
		Find(&gKs)

	knownHosts := []KnownHost{}
	for _, gK := range gKs {
		knownHosts = append(knownHosts, knownHostFromGormKnownHost(gK))
	}

	return knownHosts, count
}
