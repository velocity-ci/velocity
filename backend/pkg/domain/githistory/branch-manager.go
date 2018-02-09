package githistory

import (
	"fmt"
	"time"

	"github.com/asdine/storm"
	"github.com/velocity-ci/velocity/backend/pkg/domain"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

// Event constants
const (
	EventBranchCreate = "branch:new"
	EventBranchUpdate = "branch:update"
)

type BranchManager struct {
	db      *branchStormDB
	brokers []domain.Broker
}

func NewBranchManager(db *storm.DB) *BranchManager {
	return &BranchManager{
		db:      newBranchStormDB(db),
		brokers: []domain.Broker{},
	}
}

func (m *BranchManager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *BranchManager) Create(
	p *project.Project,
	name string,
) *Branch {
	b := &Branch{
		ID:          uuid.NewV3(uuid.NewV1(), p.ID).String(),
		Project:     p,
		Name:        name,
		LastUpdated: time.Now().UTC(),
		Active:      true,
	}
	m.db.save(b)

	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   "branches",
			Event:   EventBranchCreate,
			Payload: b,
		})
	}

	return b
}

func (m *BranchManager) Update(b *Branch) error {
	if err := m.db.save(b); err != nil {
		return err
	}
	for _, br := range m.brokers {
		br.EmitAll(&domain.Emit{
			Topic:   fmt.Sprintf("branch:%s", b.ID),
			Event:   EventBranchUpdate,
			Payload: b,
		})
	}
	return nil
}

func (m *BranchManager) GetByProjectAndName(p *project.Project, name string) (*Branch, error) {
	return m.db.getByProjectAndName(p, name)
}

func (m *BranchManager) GetAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Branch, int) {
	return m.db.getAllForProject(p, q)
}

func (m *BranchManager) GetAllForCommit(c *Commit, q *domain.PagingQuery) ([]*Branch, int) {
	return m.db.getAllForCommit(c, q)
}

func (m *BranchManager) HasCommit(b *Branch, c *Commit) bool {
	return m.db.hasCommit(b, c)
}
