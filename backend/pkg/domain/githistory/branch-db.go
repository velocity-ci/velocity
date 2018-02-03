package githistory

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type GormBranch struct {
	UUID        string `gorm:"primary_key"`
	Name        string
	Project     *project.GormProject `gorm:"ForeignKey:ProjectID"`
	ProjectID   string
	LastUpdated time.Time
	Active      bool
}

func (GormBranch) TableName() string {
	return "branches"
}

func (gB *GormBranch) ToBranch() *Branch {
	return &Branch{
		UUID:        gB.UUID,
		Name:        gB.Name,
		Project:     gB.Project.ToProject(),
		LastUpdated: gB.LastUpdated,
		Active:      gB.Active,
	}
}

func (b *Branch) ToGormBranch() *GormBranch {
	return &GormBranch{
		UUID:        b.UUID,
		Name:        b.Name,
		Project:     b.Project.ToGormProject(),
		LastUpdated: b.LastUpdated,
		Active:      b.Active,
	}
}

type branchDB struct {
	db *gorm.DB
}

func newBranchDB(gorm *gorm.DB) *branchDB {
	return &branchDB{
		db: gorm,
	}
}

func (db *branchDB) save(b *Branch) error {
	tx := db.db.Begin()

	g := b.ToGormBranch()

	tx.
		Where(GormBranch{UUID: b.UUID}).
		Assign(&g).
		FirstOrCreate(&g)

	return tx.Commit().Error
}

func (db *branchDB) getAllForProject(p *project.Project, q *domain.PagingQuery) (r []*Branch, t int) {
	t = 0

	gS := []GormBranch{}
	d := db.db

	d = d.
		Preload("Project").
		Where("project_id = ?", p.UUID)

	d.Find(&gS).Count(&t)

	d.
		Limit(q.Limit).
		Offset((q.Page - 1) * q.Limit).
		Find(&gS)

	for _, g := range gS {
		r = append(r, g.ToBranch())
	}

	return r, t
}
