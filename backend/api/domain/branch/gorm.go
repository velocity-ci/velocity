package branch

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type GORMBranch struct {
	Name      string `gorm:"primary_key"`
	ProjectID string `gorm:"primary_key"`
}

func (g GORMBranch) TableName() string {
	return "branches"
}

func gormBranchFromProjectAndBranch(p *project.Project, b *Branch) *GORMBranch {
	return &GORMBranch{
		Name:      b.Name,
		ProjectID: p.ID,
	}
}

func branchFromGORMBranch(g *GORMBranch) *Branch {
	return &Branch{
		Name: g.Name,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMBranch{})
	return &gormRepository{
		gorm: db,
	}
}

func (r *gormRepository) SaveToProject(p *project.Project, b *Branch) *Branch {
	tx := r.gorm.Begin()

	gormBranch := gormBranchFromProjectAndBranch(p, b)

	tx.
		Where(GORMBranch{Name: gormBranch.Name, ProjectID: p.ID}).
		Assign(gormBranch).
		FirstOrCreate(gormBranch)

	tx.Commit()
	return b
}

func (r *gormRepository) DeleteFromProject(p *project.Project, b *Branch) {
	tx := r.gorm.Begin()

	gormBranch := gormBranchFromProjectAndBranch(p, b)

	if err := tx.Delete(gormBranch).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetByProjectAndName(p *project.Project, name string) (*Branch, error) {
	gormBranch := &GORMBranch{}
	if r.gorm.
		Where(&GORMBranch{Name: name, ProjectID: p.ID}).
		First(gormBranch).RecordNotFound() {
		log.Printf("Could not find Branch %s", name)
		return nil, fmt.Errorf("could not find Branch %s", name)
	}

	return branchFromGORMBranch(gormBranch), nil
}

func (r *gormRepository) GetAllByProject(p *project.Project, q Query) ([]*Branch, uint64) {
	gormBranches := []GORMBranch{}
	var count uint64
	r.gorm.
		Where(&GORMBranch{ProjectID: p.ID}).
		Find(&gormBranches).
		Count(&count)

	branches := []*Branch{}
	for _, gBranch := range gormBranches {
		branches = append(branches, branchFromGORMBranch(&gBranch))
	}

	return branches, count
}
