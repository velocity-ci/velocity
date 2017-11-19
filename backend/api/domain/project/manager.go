package project

import (
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

type Manager struct {
	gormRepository *gormRepository

	Sync func(p *Project, bare bool, full bool, submodule bool, emitter velocity.Emitter) (*git.Repository, string, error)
}

func NewManager(
	db *gorm.DB,
	syncFunc func(p *Project, bare bool, full bool, submodule bool, emitter velocity.Emitter) (*git.Repository, string, error),
) *Manager {
	return &Manager{
		gormRepository: newGORMRepository(db),
		Sync:           syncFunc,
	}
}

func (m *Manager) Save(p *Project) *Project {
	m.gormRepository.Save(p)
	return p
}

func (m *Manager) Delete(p *Project) {
	m.gormRepository.Delete(p)
}

func (m *Manager) GetByID(ID string) (*Project, error) {
	return m.gormRepository.GetByID(ID)
}

func (m *Manager) GetAll(q Query) ([]*Project, uint64) {
	return m.gormRepository.GetAll(q)
}
