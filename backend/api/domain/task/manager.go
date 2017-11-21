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

func (m *Manager) Save(t *Task) *Task {
	m.gormRepository.Save(t)
	return t
}

func (m *Manager) Delete(t *Task) {
	m.gormRepository.Delete(t)
}

func (m *Manager) GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, ID string) (*Task, error) {
	return m.gormRepository.GetByProjectAndCommitAndID(p, c, ID)
}

func (m *Manager) GetAllByProjectAndCommit(p *project.Project, c *commit.Commit, q Query) ([]*Task, uint64) {
	return m.gormRepository.GetAllByProjectAndCommit(p, c, q)
}
