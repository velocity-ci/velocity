package project

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type Manager struct {
	*SyncManager
	bolt   *bolt.DB
	logger *log.Logger
}

func NewManager(m *SyncManager, bolt *bolt.DB) *Manager {
	return &Manager{
		SyncManager: m,
		bolt:        bolt,
		logger:      log.New(os.Stdout, "[project-manager]", log.Lshortfile),
	}
}

func (m *Manager) FindByID(ID string) (*Project, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}
	projectBucket := projectsBucket.Bucket([]byte(ID))
	if projectBucket == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}

	v := projectBucket.Get([]byte("info"))

	p := Project{}
	err = json.Unmarshal(v, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (m *Manager) DeleteByID(ID string) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))

	if err := projectsBucket.DeleteBucket([]byte(ID)); err != nil {
		return fmt.Errorf("Project not found: %s", ID)
	}

	return tx.Commit()
}

func (m *Manager) Save(p *Project) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	m.logger.Printf("saving project: %s", p.ID)

	projectsBucket, err := tx.CreateBucketIfNotExists([]byte("projects"))
	projectsBucket = tx.Bucket([]byte("projects"))
	projectBucket, err := projectsBucket.CreateBucketIfNotExists([]byte(p.ID))
	if err != nil {
		return err
	}
	if projectBucket == nil {
		projectBucket = projectsBucket.Bucket([]byte(p.ID))
	}

	projectJSON, err := json.Marshal(p)
	projectBucket.Put([]byte("info"), projectJSON)

	return tx.Commit()
}

func (m *Manager) FindAll() []Project {
	projects := []Project{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return projects
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	if b == nil {
		return projects
	}

	c := b.Cursor()
	for k, _ := c.First(); k != nil; k, _ = c.Next() {
		pB := b.Bucket(k)
		v := pB.Get([]byte("info"))
		p := Project{}
		err := json.Unmarshal(v, &p)
		if err == nil {
			projects = append(projects, p)
		}
	}

	return projects
}
