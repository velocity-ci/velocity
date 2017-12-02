package branch

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
)

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	logger *log.Logger
	gorm   *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(Branch{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:branch]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(b Branch) Branch {
	tx := r.gorm.Begin()

	err := tx.Where(&Branch{
		ID: b.ID,
	}).First(&Branch{}).Error
	if err != nil {
		err = tx.Create(&b).Error
	} else {
		tx.Save(&b)
	}

	tx.Commit()
	return b
}

func (r *gormRepository) Delete(b Branch) {
	tx := r.gorm.Begin()

	if err := tx.Delete(b).Error; err != nil {
		tx.Rollback()
		r.logger.Fatal(err)
	}

	tx.Commit()
}
func (r *gormRepository) GetByProjectIDAndName(projectID string, name string) (Branch, error) {
	b := Branch{}
	if r.gorm.
		Preload("Project").
		Where(&Branch{
			Name:      name,
			ProjectID: projectID,
		}).
		First(&b).RecordNotFound() {
		r.logger.Printf("Could not find Branch %s", name)
		return Branch{}, fmt.Errorf("could not find Branch %s", name)
	}

	return b, nil
}

func (r *gormRepository) GetAllByProjectID(projectID string, q Query) ([]Branch, uint64) {
	branches := []Branch{}
	var count uint64
	r.gorm.
		Preload("Project").
		Where(&Branch{
			ProjectID: projectID,
		}).
		Find(&branches).
		Count(&count)

	return branches, count
}
