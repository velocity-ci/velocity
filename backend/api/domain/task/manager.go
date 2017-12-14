package task

import (
	"github.com/velocity-ci/velocity/backend/api/websocket"

	"github.com/jinzhu/gorm"
)

type Manager struct {
	gormRepository   *gormRepository
	websocketManager *websocket.Manager
}

func NewManager(
	db *gorm.DB,
	websocketManager *websocket.Manager,
) *Manager {
	return &Manager{
		gormRepository:   newGORMRepository(db),
		websocketManager: websocketManager,
	}
}

func (m *Manager) Create(t Task) Task {
	m.gormRepository.Save(t)
	return t
}

func (m *Manager) Update(t Task) Task {
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
