package project

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/task"
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

func (m *BoltManager) GetCommitInProject(hash string, p *domain.Project) (*domain.Commit, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	b = b.Bucket([]byte(fmt.Sprintf("commits-%s", p.ID)))
	if b == nil {
		return nil, fmt.Errorf("Could not find project: %s, commit: %s", p.ID, hash)
	}

	v := b.Get([]byte(hash))
	if v == nil {
		return nil, fmt.Errorf("Could not find project: %s, commit: %s", p.ID, hash)
	}

	c := domain.Commit{}
	err = json.Unmarshal(v, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
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

func (m *BoltManager) SaveTaskForCommitInProject(t *task.Task, c *domain.Commit, p *domain.Project) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	b.CreateBucketIfNotExists([]byte(fmt.Sprintf("commits-%s", p.ID)))
	b = b.Bucket([]byte(fmt.Sprintf("commits-%s", p.ID)))
	b.CreateBucketIfNotExists([]byte(fmt.Sprintf("tasks-%s", c.Hash)))
	b = b.Bucket([]byte(fmt.Sprintf("tasks-%s", c.Hash)))
	taskJSON, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
	}
	b.Put([]byte(t.Name), taskJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	m.boltLogger.Printf("Saved task %s for %s in %s", t.Name, c.Hash, p.ID)

	return nil
}

func (m *BoltManager) GetTasksForCommitInProject(c *domain.Commit, p *domain.Project) []task.Task {
	tasks := []task.Task{}

	tx, err := m.bolt.Begin(false)
	if err != nil {
		return tasks
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte("projects"))
	b = b.Bucket([]byte(fmt.Sprintf("commits-%s", p.ID)))
	if b == nil {
		return tasks
	}
	b = b.Bucket([]byte(fmt.Sprintf("tasks-%s", c.Hash)))
	b.ForEach(func(k, v []byte) error {
		t := task.NewTask()
		err := json.Unmarshal(v, &t)
		if err != nil {
			fmt.Println(err)
		}
		tasks = append(tasks, t)
		return nil
	})

	return tasks
}
