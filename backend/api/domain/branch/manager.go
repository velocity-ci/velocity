package branch

import (
	"github.com/jinzhu/gorm"
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

func (m *Manager) Save(b Branch) Branch {
	m.gormRepository.Save(b)
	return b
}

func (m *Manager) Delete(b Branch) {
	m.gormRepository.Delete(b)
}

func (m *Manager) GetByProjectIDAndName(projectID string, name string) (Branch, error) {
	return m.gormRepository.GetByProjectIDAndName(projectID, name)
}

func (m *Manager) GetAllByProjectID(projectID string, q Query) ([]Branch, uint64) {
	return m.gormRepository.GetAllByProjectID(projectID, q)
}
