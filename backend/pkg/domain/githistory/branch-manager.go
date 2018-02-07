package githistory

import (
	"time"

	"github.com/asdine/storm"
	"github.com/velocity-ci/velocity/backend/pkg/domain"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type BranchManager struct {
	db *branchStormDB
}

func NewBranchManager(db *storm.DB) *BranchManager {
	return &BranchManager{
		db: newBranchStormDB(db),
	}
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

	return b
}

func (m *BranchManager) Update(b *Branch) error {
	return m.db.save(b)
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
