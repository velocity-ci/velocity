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
	b = tx.Bucket([]byte("projects"))
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

	b := tx.Bucket([]byte("projects"))
	if b == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}
	v := b.Get([]byte(ID))
	if v == nil {
		return nil, fmt.Errorf("Project not found: %s", ID)
	}

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
	b.ForEach(func(k, v []byte) error {
		p := domain.Project{}
		err := json.Unmarshal(v, &p)
		if err == nil {
			projects = append(projects, p)
		}
		return nil
	})

	return projects
}

func (m *BoltManager) FindAllCommitsForProject(p *domain.Project) []domain.Commit {
	commits := []domain.Commit{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return commits
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	b = b.Bucket([]byte(fmt.Sprintf("commits-%s", p.ID)))
	if b == nil {
		return commits
	}
	b.ForEach(func(k, v []byte) error {
		c := domain.Commit{}
		err := json.Unmarshal(v, &c)
		if err != nil {
			fmt.Println(err)
		}
		commits = append(commits, c)
		return nil
	})

	return commits
}

func (m *BoltManager) SaveCommitForProject(p *domain.Project, c *domain.Commit) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	b.CreateBucketIfNotExists([]byte(fmt.Sprintf("commits-%s", p.ID)))
	b = b.Bucket([]byte(fmt.Sprintf("commits-%s", p.ID)))
	commitJSON, _ := json.Marshal(c)
	b.Put([]byte(c.Hash), commitJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	m.boltLogger.Printf("Saved commit %s for %s", c.Hash, p.ID)

	return nil
}
