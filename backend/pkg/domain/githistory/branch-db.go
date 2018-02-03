package githistory

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type GormBranch struct {
	UUID        string `gorm:"primary_key"`
	Name        string
	Project     *project.GormProject
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

func (db *branchDB) delete(b *Branch) error {
	tx := db.db.Begin()

	g := b.ToGormBranch()

	if err := tx.Delete(g).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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
