package commit

import (
	"github.com/jinzhu/gorm"
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

func (m *Manager) Save(c Commit) Commit {
	m.gormRepository.Save(c)
	return c
}

func (m *Manager) Delete(c Commit) {
	m.gormRepository.Delete(c)
}

func (m *Manager) GetByProjectAndHash(p project.Project, hash string) (Commit, error) {
	return m.gormRepository.GetByProjectAndHash(p, hash)
}

func (m *Manager) GetAllByProject(p project.Project, q Query) ([]Commit, uint64) {
	return m.gormRepository.GetAllByProject(p, q)
}
