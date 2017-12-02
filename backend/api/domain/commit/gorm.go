package commit

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/branch"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type GORMCommit struct {
	ID               string `gorm:"primary_key"`
	Hash             string
	Project          project.GORMProject `gorm:"ForeignKey:ProjectReference"`
	ProjectReference string
	Author           string
	CreatedAt        time.Time
	Message          string
	Branches         []branch.GORMBranch `gorm:"many2many:commit_branches;AssociationForeignKey:ID;ForeignKey:ID"`
}

func (g GORMCommit) TableName() string {
	return "commits"
}

func GORMCommitFromCommit(c Commit) GORMCommit {
	gormBranches := []branch.GORMBranch{}
	for _, b := range c.Branches {
		gormBranches = append(gormBranches, branch.GORMBranchFromBranch(b))
	}
	return GORMCommit{
		ID:               c.ID,
		Hash:             c.Hash,
		Project:          project.GORMProjectFromProject(c.Project),
		ProjectReference: c.Project.ID,
		Author:           c.Author,
		CreatedAt:        c.CreatedAt,
		Message:          c.Message,
		Branches:         gormBranches,
	}
}

func CommitFromGORMCommit(g GORMCommit) Commit {
	branches := []branch.Branch{}
	for _, gB := range g.Branches {
		branches = append(branches, branch.BranchFromGORMBranch(gB))
	}
	return Commit{
		ID:        g.ID,
		Hash:      g.Hash,
		Author:    g.Author,
		Project:   project.ProjectFromGORMProject(g.Project),
		CreatedAt: g.CreatedAt,
		Message:   g.Message,
		Branches:  branches,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMCommit{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:commit]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(c Commit) Commit {
	tx := r.gorm.Begin()

	gormCommit := GORMCommitFromCommit(c)

	err := tx.Where(&GORMCommit{
		ID: c.ID,
	}).First(&GORMCommit{}).Error
	if err != nil {
		err = tx.Create(&gormCommit).Error
	} else {
		tx.Save(&gormCommit)
	}

	tx.Commit()
	return c
}

func (r *gormRepository) Delete(c Commit) {
	tx := r.gorm.Begin()

	gormCommit := GORMCommitFromCommit(c)

	if err := tx.Delete(gormCommit).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetByProjectAndHash(p project.Project, hash string) (Commit, error) {
	gormCommit := GORMCommit{
		Branches: []branch.GORMBranch{},
	}
	if r.gorm.
		Preload("Project").
		Where(&GORMCommit{
			ProjectReference: p.ID,
			Hash:             hash,
		}).
		Preload("Branches").
		Preload("Branches.Project").
		First(&gormCommit).RecordNotFound() {
		r.logger.Printf("Could not find Commit %s", hash)
		return Commit{}, fmt.Errorf("could not find Commit %s", hash)
	}

	return CommitFromGORMCommit(gormCommit), nil
}

func (r *gormRepository) GetAllByProject(p project.Project, q Query) ([]Commit, uint64) {
	gormCommits := []GORMCommit{}
	var count uint64
	db := r.gorm

	db = db.
		Preload("Project").
		Where("commits.project_reference = ?", p.ID).
		Preload("Branches").
		Preload("Branches.Project")

	if len(q.Branch) > 0 {
		db = db.
			Joins("JOIN commit_branches AS cb ON cb.gorm_commit_id=commits.id").
			Joins("JOIN branches AS b ON b.id=cb.gorm_branch_id").
			Where("b.name in (?)", []string{q.Branch}).
			Group("commits.id")
	}
	db.Find(&gormCommits).Count(&count)

	db.
		Limit(int(q.Amount)).
		Offset(int(q.Page - 1)).
		Find(&gormCommits)

	commits := []Commit{}
	for _, gCommit := range gormCommits {
		commits = append(commits, CommitFromGORMCommit(gCommit))
	}

	return commits, count
}
