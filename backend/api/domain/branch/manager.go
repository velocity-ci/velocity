package branch

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

func (m *Manager) Save(b *Branch) *Branch {
	m.gormRepository.Save(b)
	return b
}

func (m *Manager) Delete(b *Branch) {
	m.gormRepository.Delete(b)
}

func (m *Manager) GetByProjectAndName(p *project.Project, name string) (*Branch, error) {
	return m.gormRepository.GetByProjectAndName(p, name)
}

func (m *Manager) GetAllByProject(p *project.Project, q Query) ([]*Branch, uint64) {
	return m.gormRepository.GetAllByProject(p, q)
}
