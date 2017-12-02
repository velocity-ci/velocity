package project

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
)

// type Project struct {
// 	ID string `gorm:"primary_key"`

// 	Name       string
// 	Repository []byte // Store as JSON

// 	CreatedAt time.Time
// 	UpdatedAt time.Time

// 	Synchronising bool
// }

// func ProjectFromProject(p Project) Project {
// 	jsonRepo, err := json.Marshal(p.Repository)
// 	if err != nil {
// 		log.Println("Could not marshal repository")
// 		log.Fatal(err)
// 	}
// 	return Project{
// 		ID:            p.ID,
// 		Name:          p.Name,
// 		Repository:    jsonRepo,
// 		CreatedAt:     p.CreatedAt,
// 		UpdatedAt:     p.UpdatedAt,
// 		Synchronising: p.Synchronising,
// 	}
// }

// func ProjectFromProject(g Project) Project {
// 	var repo velocity.GitRepository
// 	err := json.Unmarshal(g.Repository, &repo)
// 	if err != nil {
// 		log.Println("Could not unmarshal repository")
// 		log.Fatal(err)
// 	}
// 	return Project{
// 		ID:            g.ID,
// 		Name:          g.Name,
// 		Repository:    repo,
// 		CreatedAt:     g.CreatedAt,
// 		UpdatedAt:     g.UpdatedAt,
// 		Synchronising: g.Synchronising,
// 	}
// }

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(Project{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:project]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(p Project) Project {
	tx := r.gorm.Begin()

	err := tx.Where(&Project{
		ID: p.ID,
	}).First(&Project{}).Error
	if err != nil {
		err = tx.Create(&p).Error
	} else {
		tx.Save(&p)
	}

	tx.Commit()
	return p
}

func (r *gormRepository) Delete(p Project) {
	tx := r.gorm.Begin()

	if err := tx.Delete(p).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByID(ID string) (Project, error) {
	project := Project{}
	if r.gorm.Where(&Project{ID: ID}).First(&project).RecordNotFound() {
		r.logger.Printf("Could not find Project %s", ID)
		return Project{}, fmt.Errorf("could not find Project %s", ID)
	}

	return project, nil
}

func (r *gormRepository) GetAll(q Query) ([]Project, uint64) {
	projects := []Project{}
	var count uint64
	r.gorm.
		Find(&projects).
		Count(&count)

	return projects, count
}
