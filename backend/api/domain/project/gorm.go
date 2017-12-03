package project

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type gormProject struct {
	ID string `gorm:"primary_key"`

	Name       string
	Repository []byte // Store as JSON

	CreatedAt time.Time
	UpdatedAt time.Time

	Synchronising bool
}

func (gormProject) TableName() string {
	return "projects"
}

func gormProjectFromProject(p Project) gormProject {
	jsonRepo, err := json.Marshal(p.Repository)
	if err != nil {
		log.Println("Could not marshal repository")
		log.Fatal(err)
	}
	return gormProject{
		ID:            p.ID,
		Name:          p.Name,
		Repository:    jsonRepo,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		Synchronising: p.Synchronising,
	}
}

func projectFromGormProject(g gormProject) Project {
	var repo velocity.GitRepository
	err := json.Unmarshal(g.Repository, &repo)
	if err != nil {
		log.Println("Could not unmarshal repository")
		log.Fatal(err)
	}
	return Project{
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
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(gormProject{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:project]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(p Project) Project {
	tx := r.gorm.Begin()

	gP := gormProjectFromProject(p)

	err := tx.Where(&gormProject{
		ID: p.ID,
	}).First(&gormProject{}).Error
	if err != nil {
		err = tx.Create(&gP).Error
	} else {
		tx.Save(&gP)
	}

	tx.Commit()
	return projectFromGormProject(gP)
}

func (r *gormRepository) Delete(p Project) {
	tx := r.gorm.Begin()
	gP := gormProjectFromProject(p)

	if err := tx.Delete(gP).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByID(ID string) (Project, error) {
	gP := gormProject{}

	if r.gorm.Where(&gormProject{ID: ID}).First(&gP).RecordNotFound() {
		r.logger.Printf("Could not find Project %s", ID)
		return Project{}, fmt.Errorf("could not find Project %s", ID)
	}

	return projectFromGormProject(gP), nil
}

func (r *gormRepository) GetAll(q Query) ([]Project, uint64) {
	gPs := []gormProject{}
	var count uint64
	r.gorm.
		Find(&gPs).
		Count(&count)

	projects := []Project{}
	for _, gP := range gPs {
		projects = append(projects, projectFromGormProject(gP))
	}

	return projects, count
}
