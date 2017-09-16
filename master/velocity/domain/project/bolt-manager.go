package project

import (
	"encoding/json"
	"fmt"
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

	b, err := tx.CreateBucketIfNotExists([]byte("projects"))
	projectJSON, err := json.Marshal(p)
	b.Put([]byte(p.ID), projectJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *BoltManager) FindByID(ID string) (*domain.Project, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("projects"))
	v := b.Get([]byte(ID))
	if v != nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}

	p := domain.Project{}
	err = json.Unmarshal(v, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (m *BoltManager) FindAll() []*domain.Project {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return []*domain.Project{}
	}
	defer tx.Rollback()

	b, err := tx.CreateBucketIfNotExists([]byte("projects"))
	err = b.ForEach(func(k, v []byte) error {

	})
}
