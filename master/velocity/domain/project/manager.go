package project

import (
	"fmt"
	"log"

	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/task"
)

type Manager struct {
	logger      *log.Logger
	dbManager   *DBManager
	boltManager *BoltManager
}

func NewManager(
	logger *log.Logger,
	dbManager *DBManager,
	boltManager *BoltManager,
) *Manager {
	return &Manager{
		logger:      logger,
		dbManager:   dbManager,
		boltManager: boltManager,
	}
}

func (m *Manager) Save(p *domain.Project) error {
	m.boltManager.Save(p)
	m.dbManager.Save(p)
	return nil
}

func (m *Manager) FindByID(ID string) (*domain.Project, error) {
	p, err := m.boltManager.FindByID(ID)
	if err == nil {
		return p, nil
	}
	fmt.Println(err)

	p, err = m.dbManager.FindByID(ID)
	if err == nil {
		m.boltManager.Save(p)
		return p, nil
	}

	return nil, err
}

func (m *Manager) FindAll() []domain.Project {
	projects := m.boltManager.FindAll()

	if len(projects) < 1 {
		projects = m.dbManager.FindAll()
		for _, p := range projects {
			m.boltManager.Save(&p)
		}
	}
	return projects
}

func (m *Manager) GetCommitInProject(hash string, p *domain.Project) (*domain.Commit, error) {
	return m.boltManager.GetCommitInProject(hash, p)
}

func (m *Manager) GetCommitsForProject(p *domain.Project) []domain.Commit {
	commits := m.boltManager.FindAllCommitsForProject(p)

	return commits
}

func (m *Manager) SaveCommitForProject(p *domain.Project, c *domain.Commit) error {
	return m.boltManager.SaveCommitForProject(p, c)
}

func (m *Manager) SaveTaskForCommitInProject(t *task.Task, c *domain.Commit, p *domain.Project) error {
	return m.boltManager.SaveTaskForCommitInProject(t, c, p)
}

func (m *Manager) GetTasksForCommitInProject(c *domain.Commit, p *domain.Project) []task.Task {
	return m.boltManager.GetTasksForCommitInProject(c, p)
}
