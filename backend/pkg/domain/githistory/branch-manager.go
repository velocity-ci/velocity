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

func (m *BranchManager) New(
	p *project.Project,
	name string,
) *Branch {
	return &Branch{
		UUID:        uuid.NewV3(uuid.NewV1(), p.UUID).String(),
		Project:     p,
		Name:        name,
		LastUpdated: time.Now().UTC(),
		Active:      true,
	}
}

func (m *BranchManager) Save(b *Branch) error {
	return m.db.save(b)
}

func (m *BranchManager) SaveCommitToBranch(c *Commit, b *Branch) error {
	return m.db.saveCommitToBranch(c, b)
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
