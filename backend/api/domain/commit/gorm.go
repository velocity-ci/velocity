package commit

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

type gormCommit struct {
	ID        string `gorm:"primary_key"`
	ProjectID string
	Hash      string
	Author    string
	CreatedAt time.Time
	Message   string
	Branches  []gormBranch `gorm:"many2many:commit_branches;AssociationForeignKey:ID;ForeignKey:ID"`
}

func (gormCommit) TableName() string {
	return "commits"
}

func commitFromGormCommit(gC gormCommit) Commit {
	branches := []Branch{}
	for _, gB := range gC.Branches {
		branches = append(branches, branchFromGormBranch(gB))
	}

	return Commit{
		ID:        gC.ID,
		ProjectID: gC.ProjectID,
		Hash:      gC.Hash,
		Author:    gC.Author,
		CreatedAt: gC.CreatedAt,
		Message:   gC.Message,
		Branches:  branches,
	}
}

func gormCommitFromCommit(c Commit) gormCommit {
	gormBranches := []gormBranch{}
	for _, b := range c.Branches {
		gormBranches = append(gormBranches, gormBranchFromBranch(b))
	}

	return gormCommit{
		ID:        c.ID,
		ProjectID: c.ProjectID,
		Hash:      c.Hash,
		Author:    c.Author,
		CreatedAt: c.CreatedAt,
		Message:   c.Message,
		Branches:  gormBranches,
	}
}

type gormBranch struct {
	ID          string `gorm:"primary_key"`
	Name        string
	ProjectID   string
	LastUpdated time.Time
}

func (gormBranch) TableName() string {
	return "branches"
}

func branchFromGormBranch(gB gormBranch) Branch {
	return Branch{
		ID:          gB.ID,
		Name:        gB.Name,
		ProjectID:   gB.ProjectID,
		LastUpdated: gB.LastUpdated,
	}
}

func gormBranchFromBranch(b Branch) gormBranch {
	return gormBranch{
		ID:          b.ID,
		Name:        b.Name,
		ProjectID:   b.ProjectID,
		LastUpdated: b.LastUpdated,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(gormCommit{}, gormBranch{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:commit]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) SaveCommit(c Commit) Commit {
	tx := r.gorm.Begin()

	gC := gormCommitFromCommit(c)

	err := tx.Where(&gormCommit{
		ID: c.ID,
	}).First(&gormCommit{}).Error
	if err != nil {
		err = tx.Create(&gC).Error
	} else {
		tx.Save(&gC)
	}

	tx.Commit()
	r.logger.Printf("saved commit %s", c.ID)

	return commitFromGormCommit(gC)
}

func (r *gormRepository) DeleteCommit(c Commit) {
	tx := r.gorm.Begin()

	gC := gormCommitFromCommit(c)

	if err := tx.Delete(gC).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetCommitByCommitID(id string) (Commit, error) {
	gC := gormCommit{}
	if r.gorm.
		Preload("Branches").
		Where(&gormCommit{
			ID: id,
		}).
		First(&gC).RecordNotFound() {
		r.logger.Printf("could not find commit %s", id)
		return Commit{}, fmt.Errorf("could not find commit %s", id)
	}

	r.logger.Printf("got commit %s", id)
	return commitFromGormCommit(gC), nil
}

func (r *gormRepository) GetCommitByProjectIDAndCommitHash(projectID string, hash string) (Commit, error) {
	gC := gormCommit{}
	if r.gorm.
		Preload("Branches").
		Where(&gormCommit{
			ProjectID: projectID,
			Hash:      hash,
		}).
		First(&gC).RecordNotFound() {
		r.logger.Printf("could not find Project Commit %s:%s", projectID, hash)
		return Commit{}, fmt.Errorf("could not find Project Commit %s:%s", projectID, hash)
	}

	r.logger.Printf("got project:commit %s:%s", projectID, hash)
	return commitFromGormCommit(gC), nil
}

func (r *gormRepository) GetAllCommitsByProjectID(projectID string, q Query) ([]Commit, uint64) {
	gCs := []gormCommit{}
	var count uint64
	db := r.gorm

	db = db.
		Preload("Branches").
		Where("commits.project_id = ?", projectID)

	if len(q.Branch) > 0 {
		db = db.
			Joins("JOIN commit_branches AS cb ON cb.commit_id=commits.id").
			Joins("JOIN branches AS b ON b.id=cb.branch_id").
			Where("b.name in (?)", []string{q.Branch}).
			Group("commits.id")
	}
	db.Find(&gCs).Count(&count)

	db.
		Limit(int(q.Amount)).
		Offset(int(q.Page - 1)).
		Find(&gCs)

	commits := []Commit{}
	for _, gC := range gCs {
		commits = append(commits, commitFromGormCommit(gC))
	}

	return commits, count
}

func (r *gormRepository) SaveBranch(b Branch) Branch {
	tx := r.gorm.Begin()

	gB := gormBranchFromBranch(b)

	err := tx.Where(&gormBranch{
		ID: b.ID,
	}).First(&gormBranch{}).Error
	if err != nil {
		err = tx.Create(&gB).Error
	} else {
		tx.Save(&gB)
	}

	tx.Commit()
	r.logger.Printf("saved branch %s", b.ID)
	return branchFromGormBranch(gB)
}

func (r *gormRepository) DeleteBranch(b Branch) {
	tx := r.gorm.Begin()

	gB := gormBranchFromBranch(b)

	if err := tx.Delete(gB).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetBranchByProjectIDAndName(projectID string, name string) (Branch, error) {
	gB := gormBranch{}
	if r.gorm.
		Where(&gormBranch{
			Name:      name,
			ProjectID: projectID,
		}).
		First(&gB).RecordNotFound() {
		r.logger.Printf("could not find project:branch %s:%s", projectID, name)
		return Branch{}, fmt.Errorf("could not find project:branch %s:%s", projectID, name)
	}

	r.logger.Printf("got project:branch %s:%s", projectID, name)
	return branchFromGormBranch(gB), nil
}

func (r *gormRepository) GetAllBranchesByProjectID(projectID string, q Query) ([]Branch, uint64) {
	gBs := []gormBranch{}
	var count uint64
	r.gorm.
		Where(&gormBranch{
			ProjectID: projectID,
		}).
		Find(&gBs).
		Count(&count)

	branches := []Branch{}
	for _, gB := range gBs {
		branches = append(branches, branchFromGormBranch(gB))
	}

	return branches, count
}
