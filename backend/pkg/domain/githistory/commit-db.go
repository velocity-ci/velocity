package githistory

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type GormCommit struct {
	UUID      string               `gorm:"primary_key"`
	Project   *project.GormProject `gorm:"ForeignKey:ProjectID"`
	ProjectID string
	Hash      string
	Author    string
	CreatedAt time.Time
	Message   string
	Branches  []*GormBranch `gorm:"many2many:commit_branches;AssociationForeignKey:UUID;ForeignKey:UUID"`
}

func (GormCommit) TableName() string {
	return "commits"
}

func (gC *GormCommit) ToCommit() *Commit {
	branches := []*Branch{}
	for _, gB := range gC.Branches {
		branches = append(branches, gB.ToBranch())
	}
	return &Commit{
		UUID:      gC.UUID,
		Project:   gC.Project.ToProject(),
		Hash:      gC.Hash,
		Author:    gC.Author,
		CreatedAt: gC.CreatedAt,
		Message:   gC.Message,
		Branches:  branches,
	}
}

func (c *Commit) ToGormCommit() *GormCommit {
	gormBranches := []*GormBranch{}
	for _, b := range c.Branches {
		gormBranches = append(gormBranches, b.ToGormBranch())
	}
	return &GormCommit{
		UUID:      c.UUID,
		Project:   c.Project.ToGormProject(),
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  gormBranches,
	}
}

type commitDB struct {
	db *gorm.DB
}

func newCommitDB(gorm *gorm.DB) *commitDB {
	return &commitDB{
		db: gorm,
	}
}

func (db *commitDB) delete(c *Commit) error {
	tx := db.db.Begin()

	g := c.ToGormCommit()

	if err := tx.Delete(g).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (db *commitDB) save(c *Commit) error {
	tx := db.db.Begin()

	g := c.ToGormCommit()

	tx.
		Where(GormCommit{UUID: c.UUID}).
		Assign(&g).
		FirstOrCreate(&g)

	return tx.Commit().Error
}

func (db *commitDB) getByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	g := GormCommit{}
	if db.db.
		Preload("Branches").
		Preload("Project").
		Where("project_id = ? AND hash = ?", p.UUID, hash).
		First(&g).RecordNotFound() {
		return nil, fmt.Errorf("could not find project:commit %s:%s", p.UUID, hash)
	}
	return g.ToCommit(), nil
}
