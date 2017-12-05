package commit

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/websocket"
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

func (m *Manager) CreateCommit(c Commit) Commit {
	m.gormRepository.SaveCommit(c)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", c.ProjectID),
		Event:   websocket.VNewCommit,
		Payload: NewResponseCommit(c),
	})

	return c
}

func (m *Manager) UpdateCommit(c Commit) Commit {
	m.gormRepository.SaveCommit(c)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", c.ProjectID),
		Event:   websocket.VUpdateCommit,
		Payload: NewResponseCommit(c),
	})
	return c
}

func (m *Manager) DeleteCommit(c Commit) {
	m.gormRepository.DeleteCommit(c)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", c.ProjectID),
		Event:   websocket.VDeleteCommit,
		Payload: NewResponseCommit(c),
	})
}

func (m *Manager) GetCommitByCommitID(id string) (Commit, error) {
	return m.gormRepository.GetCommitByCommitID(id)
}

func (m *Manager) GetCommitByProjectIDAndCommitHash(projectID string, hash string) (Commit, error) {
	return m.gormRepository.GetCommitByProjectIDAndCommitHash(projectID, hash)
}

func (m *Manager) GetAllCommitsByProjectID(projectID string, q Query) ([]Commit, uint64) {
	return m.gormRepository.GetAllCommitsByProjectID(projectID, q)
}

func (m *Manager) CreateBranch(b Branch) Branch {
	b.LastUpdated = time.Now()
	m.gormRepository.SaveBranch(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VNewBranch,
		Payload: NewResponseBranch(b),
	})
	return b
}

func (m *Manager) UpdateBranch(b Branch) Branch {
	b.LastUpdated = time.Now()
	m.gormRepository.SaveBranch(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VUpdateBranch,
		Payload: NewResponseBranch(b),
	})
	return b
}

func (m *Manager) DeleteBranch(b Branch) {
	m.gormRepository.DeleteBranch(b)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", b.ProjectID),
		Event:   websocket.VDeleteBranch,
		Payload: NewResponseBranch(b),
	})
}

func (m *Manager) GetBranchByProjectIDAndName(projectID string, name string) (Branch, error) {
	return m.gormRepository.GetBranchByProjectIDAndName(projectID, name)
}

func (m *Manager) GetAllBranchesByProjectID(projectID string, q Query) ([]Branch, uint64) {
	return m.gormRepository.GetAllBranchesByProjectID(projectID, q)
}
