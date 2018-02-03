package githistory

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type BranchManager struct {
	db *branchDB
}

func NewBranchManager(db *gorm.DB) *BranchManager {
	db.AutoMigrate(&GormCommit{}, &GormBranch{})
	return &BranchManager{
		db: newBranchDB(db),
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

func (m *BranchManager) GetAllForProject(p *project.Project, q *domain.PagingQuery) ([]*Branch, int) {
	return m.db.getAllForProject(p, q)
}
