package branch

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
)

type GORMBranch struct {
	ID               string `gorm:"primary_key"`
	Name             string
	Project          project.GORMProject `gorm:"ForeignKey:ProjectReference"`
	ProjectReference string
}

func (g GORMBranch) TableName() string {
	return "branches"
}

func GORMBranchFromBranch(b *Branch) *GORMBranch {
	return &GORMBranch{
		ID:               b.ID,
		Name:             b.Name,
		Project:          *project.GORMProjectFromProject(&b.Project),
		ProjectReference: b.Project.ID,
	}
}

func BranchFromGORMBranch(g *GORMBranch) *Branch {
	return &Branch{
		ID:      g.ID,
		Name:    g.Name,
		Project: *project.ProjectFromGORMProject(&g.Project),
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

func (r *gormRepository) Save(b *Branch) *Branch {
	tx := r.gorm.Begin()

	gormBranch := GORMBranchFromBranch(b)

	tx.
		Where(GORMBranch{ID: gormBranch.ID}).
		Assign(gormBranch).
		FirstOrCreate(gormBranch)

	tx.Commit()
	return b
}

func (r *gormRepository) Delete(b *Branch) {
	tx := r.gorm.Begin()

	gormBranch := GORMBranchFromBranch(b)

	if err := tx.Delete(gormBranch).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetByProjectAndName(p *project.Project, name string) (*Branch, error) {
	gormBranch := &GORMBranch{}
	if r.gorm.
		Preload("Project").
		Where(&GORMBranch{
			Name:             name,
			ProjectReference: p.ID,
		}).
		First(gormBranch).RecordNotFound() {
		log.Printf("Could not find Branch %s", name)
		return nil, fmt.Errorf("could not find Branch %s", name)
	}

	return BranchFromGORMBranch(gormBranch), nil
}

func (r *gormRepository) GetAllByProject(p *project.Project, q Query) ([]*Branch, uint64) {
	gormBranches := []GORMBranch{}
	var count uint64
	r.gorm.
		Preload("Project").
		Where(&GORMBranch{
			ProjectReference: p.ID,
		}).
		Find(&gormBranches).
		Count(&count)

	branches := []*Branch{}
	for _, gBranch := range gormBranches {
		branches = append(branches, BranchFromGORMBranch(&gBranch))
	}

	return branches, count
}
