package project

import (
	"log"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type Manager struct {
	logger      *log.Logger
	dbManager   *DBManager
	boltManager *BoltManager
}

func NewManager(
	logger *log.Logger,
	dbManager *DBManager,
	boltManager *BoltManager,
) *Manager {
	return &Manager{
		logger:      logger,
		dbManager:   dbManager,
		boltManager: boltManager,
	}
}

func (m *Manager) Save(p *domain.Project) error {
	m.boltManager.Save(p)
	m.dbManager.Save(p)
	return nil
}

func (m *Manager) FindByID(ID string) (*domain.Project, error) {
	p, err := m.boltManager.FindByID(ID)
	if err == nil {
		return p, nil
	}

	p, err = m.dbManager.FindByID(ID)
	if err == nil {
		m.boltManager.Save(p)
		return p, nil
	}

	return nil, err
}

func (m *Manager) FindAll() []*domain.Project {

}
