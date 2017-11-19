package commit

import (
	"fmt"
	"log"
	"time"

	"github.com/velocity-ci/velocity/backend/api/domain/branch"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type GORMCommit struct {
	Hash      string `gorm:"primary_key"`
	ProjectID string
	Author    string
	CreatedAt time.Time
	Message   string
	Branches  []branch.GORMBranch `gorm:"many2many:commit_branches;AssociationForeignKey:Name;ForeignKey:Hash"`
}

func (g GORMCommit) TableName() string {
	return "commits"
}

func gormCommitFromProjectAndCommit(p *project.Project, c *Commit) *GORMCommit {
	gormBranches := []branch.GORMBranch{}
	for _, b := range c.Branches {
		gormBranches = append(gormBranches, branch.GORMBranch{Name: b.Name, ProjectID: p.ID})
	}
	return &GORMCommit{
		Hash:      c.Hash,
		ProjectID: p.ID,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  gormBranches,
	}
}

func commitFromGORMCommit(g *GORMCommit) *Commit {
	branches := []branch.Branch{}
	for _, gB := range g.Branches {
		branches = append(branches, branch.Branch{Name: gB.Name})
	}
	return &Commit{
		Hash:      g.Hash,
		Author:    g.Author,
		CreatedAt: g.CreatedAt,
		Message:   g.Message,
		Branches:  branches,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMCommit{})
	return &gormRepository{
		gorm: db,
	}
}

func (r *gormRepository) SaveToProject(p *project.Project, c *Commit) *Commit {
	tx := r.gorm.Begin()

	gormCommit := gormCommitFromProjectAndCommit(p, c)

	err := tx.Where("hash = ? AND project_id = ?", c.Hash, p.ID).First(&GORMCommit{}).Error
	if err != nil {
		err = tx.Create(gormCommit).Error
	} else {
		tx.Save(gormCommit)
	}

	// tx.
	// 	Where(GORMCommit{Hash: c.Hash, ProjectID: p.ID}).
	// 	Assign(gormCommit).
	// 	FirstOrCreate(gormCommit)

	tx.Commit()
	return c
}

func (r *gormRepository) DeleteFromProject(p *project.Project, c *Commit) {
	tx := r.gorm.Begin()

	gormCommit := gormCommitFromProjectAndCommit(p, c)

	if err := tx.Delete(gormCommit).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetByProjectAndHash(p *project.Project, hash string) (*Commit, error) {
	gormCommit := GORMCommit{
		Branches: []branch.GORMBranch{},
	}
	if r.gorm.
		Where(&project.GORMProject{ID: p.ID}).
		Where(&GORMCommit{Hash: hash}).
		Related(&gormCommit.Branches, "Branches").
		First(&gormCommit).RecordNotFound() {
		log.Printf("Could not find Commit %s", hash)
		return nil, fmt.Errorf("could not find Commit %s", hash)
	}

	return commitFromGORMCommit(&gormCommit), nil
}

func (r *gormRepository) GetAllByProject(p *project.Project, q Query) ([]*Commit, uint64) {
	gormCommits := []GORMCommit{}
	var count uint64
	db := r.gorm

	log.Println(q)

	db = db.
		Where("commits.project_id = ?", p.ID)

	if len(q.Branch) > 0 {
		db = db.
			Joins("JOIN commit_branches ON commits.hash=commit_branches.gorm_commit_hash").
			Joins("JOIN branches ON commit_branches.gorm_branch_name=branches.name").
			Where("branches.project_id = ?", p.ID).
			Where("branches.name in (?)", []string{q.Branch}).
			Group("commits.hash")
	}
	db.Find(&gormCommits).Count(&count)

	db.
		Limit(int(q.Amount)).
		Offset(int(q.Page - 1)).
		Find(&gormCommits)

	commits := []*Commit{}
	for _, gCommit := range gormCommits {
		gCommit.Branches = []branch.GORMBranch{}
		r.gorm.Model(&gCommit).Related(&gCommit.Branches, "Branches")
		commits = append(commits, commitFromGORMCommit(&gCommit))
	}

	return commits, count
}
