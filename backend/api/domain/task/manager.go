package task

import (
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type Manager struct {
	gormRepository *gormRepository
}

func NewManager(
	db *gorm.DB,
) *Manager {
	return &Manager{
		gormRepository: newGORMRepository(db),
	}
}

func (m *Manager) SaveToProjectAndCommit(p *project.Project, c *commit.Commit, t *Task) *Task {
	m.gormRepository.SaveToProjectAndCommit(p, c, t)
	return t
}

func (m *Manager) DeleteFromProjectAndCommit(p *project.Project, c *commit.Commit, t *Task) {
	m.gormRepository.DeleteFromProjectAndCommit(p, c, t)
}

func (m *Manager) GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, ID string) (*Task, error) {
	return m.gormRepository.GetByProjectAndCommitAndID(p, c, ID)
}

func (m *Manager) GetAllByProjectAndCommit(p *project.Project, c *commit.Commit, q Query) ([]*Task, uint64) {
	return m.gormRepository.GetAllByProjectAndCommit(p, c, q)
}
