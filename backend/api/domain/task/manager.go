package task

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

func (m *Manager) Save(t Task) Task {
	m.gormRepository.Save(t)
	return t
}

func (m *Manager) Delete(t Task) {
	m.gormRepository.Delete(t)
}

func (m *Manager) GetByTaskID(ID string) (Task, error) {
	return m.gormRepository.GetByTaskID(ID)
}

func (m *Manager) GetByCommitIDAndTaskName(commitID string, name string) (Task, error) {
	return m.gormRepository.GetByCommitIDAndTaskName(commitID, name)
}

func (m *Manager) GetAllByCommitID(commitID string, q Query) ([]Task, uint64) {
	return m.gormRepository.GetAllByCommitID(commitID, q)
}
