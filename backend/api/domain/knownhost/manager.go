package knownhost

import (
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/websocket"
)

type Manager struct {
	gormRepository   *gormRepository
	websocketManager *websocket.Manager
	fileManager      *fileManager
}

func NewManager(
	db *gorm.DB,
	websocketManager *websocket.Manager,
) *Manager {
	return &Manager{
		gormRepository:   newGORMRepository(db),
		websocketManager: websocketManager,
		fileManager:      NewFileManager(),
	}
}

func (m *Manager) Create(k KnownHost) KnownHost {
	m.gormRepository.Save(k)
	m.fileManager.Save(k)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   "knownhosts",
		Event:   websocket.VNewBranch,
		Payload: NewResponseKnownHost(k),
	})

	return k
}

func (m *Manager) Delete(k KnownHost) {
	m.gormRepository.Delete(k)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   "knownhosts",
		Event:   websocket.VDeleteCommit,
		Payload: NewResponseKnownHost(k),
	})
}

func (m *Manager) GetByID(id string) (KnownHost, error) {
	return m.gormRepository.GetByID(id)
}

func (m *Manager) GetAll(q KnownHostQuery) ([]KnownHost, uint64) {
	return m.gormRepository.GetAll(q)
}

func (m *Manager) Exists(entry string) bool {
	return m.fileManager.Exists(entry)
}

func (m *Manager) GetAllEntries() []string {
	ks, _ := m.gormRepository.GetAll(KnownHostQuery{})
	entries := []string{}
	for _, k := range ks {
		entries = append(entries, k.Entry)
	}

	return entries
}
