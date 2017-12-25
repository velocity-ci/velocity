package project

import (
	"fmt"
	"io"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/websocket"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

type Manager struct {
	gormRepository   *gormRepository
	websocketManager *websocket.Manager

	Sync func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error)
}

func NewManager(
	db *gorm.DB,
	syncFunc func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error),
	websocketManager *websocket.Manager,
) *Manager {
	return &Manager{
		gormRepository:   newGORMRepository(db),
		Sync:             syncFunc,
		websocketManager: websocketManager,
	}
}

func (m *Manager) Create(p Project) Project {
	m.gormRepository.Save(p)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   "projects",
		Event:   websocket.VNewProject,
		Payload: NewResponseProject(p),
	})
	return p
}

func (m *Manager) Update(p Project) Project {
	p.UpdatedAt = time.Now()
	m.gormRepository.Save(p)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic:   fmt.Sprintf("project:%s", p.ID),
		Event:   websocket.VUpdateProject,
		Payload: NewResponseProject(p),
	})
	return p
}

func (m *Manager) Delete(p Project) {
	m.gormRepository.Delete(p)
	m.websocketManager.EmitAll(&websocket.PhoenixMessage{
		Topic: fmt.Sprintf("project:%s", p.ID),
		Event: websocket.VDeleteProject,
	})
}

func (m *Manager) GetByID(ID string) (Project, error) {
	return m.gormRepository.GetByID(ID)
}

func (m *Manager) GetAll(q Query) ([]Project, uint64) {
	return m.gormRepository.GetAll(q)
}
