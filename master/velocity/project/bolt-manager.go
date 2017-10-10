package project

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type BoltManager struct {
	logger *log.Logger
	bolt   *bolt.DB
}

func NewBoltManager(
	bolt *bolt.DB,
) *BoltManager {
	return &BoltManager{
		logger: log.New(os.Stdout, "[bolt-project]", log.Lshortfile),
		bolt:   bolt,
	}
}

func (m *BoltManager) Save(p *domain.Project) error {
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

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
func (m *BoltManager) DeleteById(ID string) error {

	tx, err := m.bolt.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	projectsBucket := tx.Bucket([]byte("projects"))

	if err := projectsBucket.DeleteBucket([]byte(ID)); err != nil {
		return fmt.Errorf("Project not found: %s", ID)
	}

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

	projectsBucket := tx.Bucket([]byte("projects"))
	if projectsBucket == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}
	projectBucket := projectsBucket.Bucket([]byte(ID))
	if projectBucket == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}

	v := projectBucket.Get([]byte("info"))

	p := domain.Project{}
	err = json.Unmarshal(v, &p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (m *BoltManager) FindAll() []domain.Project {
	projects := []domain.Project{}

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
		p := domain.Project{}
		err := json.Unmarshal(v, &p)
		if err == nil {
			projects = append(projects, p)
		}
	}

	return projects
}
