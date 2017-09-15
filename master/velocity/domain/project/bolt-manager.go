package project

import (
	"encoding/json"
	"log"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type BoltManager struct {
	boltLogger *log.Logger
	bolt       *bolt.DB
}

func NewBoltManager(
	boltLogger *log.Logger,
	bolt *bolt.DB,
) *BoltManager {
	return &BoltManager{
		boltLogger: boltLogger,
		bolt:       bolt,
	}
}

func (m *BoltManager) Save(p *domain.Project) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte(p.ID))
	projectJSON, err := json.Marshal(p)
	b.Put([]byte(p.ID), projectJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
