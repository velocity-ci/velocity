package project

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

type GORMProject struct {
	ID string `gorm:"primary_key"`

	Name       string
	Repository []byte // Store as JSON

	CreatedAt time.Time
	UpdatedAt time.Time

	Synchronising bool
}

func gormProjectFromProject(p *Project) *GORMProject {
	jsonRepo, err := json.Marshal(p.Repository)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return &GORMProject{
		ID:            p.ID,
		Name:          p.Name,
		Repository:    jsonRepo,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
}

func projectFromGORMProject(g *GORMProject) *Project {
	var repo GitRepository
	err := json.Unmarshal(g.Repository, &repo)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return &Project{
		ID:            g.ID,
		Name:          g.Name,
		Repository:    repo,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
		Synchronising: g.Synchronising,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMProject{})
	return &gormRepository{
		gorm: db,
	}
}

func (r *gormRepository) Save(p *Project) *Project {
	tx := r.gorm.Begin()

	gormProject := gormProjectFromProject(p)

	tx.
		Where(GORMProject{ID: p.ID}).
		Assign(gormProject).
		FirstOrCreate(gormProject)

	tx.Commit()
	return p
}

func (r *gormRepository) Delete(p *Project) {
	tx := r.gorm.Begin()

	gormProject := gormProjectFromProject(p)

	if err := tx.Delete(gormProject).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByID(ID string) (*Project, error) {
	gormProject := &GORMProject{}
	if r.gorm.Where(&GORMProject{ID: ID}).First(gormProject).RecordNotFound() {
		log.Printf("Could not find Project %s", ID)
		return nil, fmt.Errorf("could not find Project %s", ID)
	}

	return projectFromGORMProject(gormProject), nil
}

func (r *gormRepository) GetAll(q Query) ([]*Project, uint64) {
	gormProjects := []GORMProject{}
	var count uint64
	r.gorm.
		Find(&gormProjects).
		Count(&count)

	projects := []*Project{}
	for _, gProject := range gormProjects {
		projects = append(projects, projectFromGORMProject(&gProject))
	}

	return projects, count
}
