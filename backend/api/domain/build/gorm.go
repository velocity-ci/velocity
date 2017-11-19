package build

import "github.com/jinzhu/gorm"

type GORMBuild struct {
	ID         string `gorm:"primary_key"`
	ProjectID  string
	CommitHash string
	TaskID     string
	Status     string
	Parameters []byte // Parameters as JSON
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMBuild{})
	return &gormRepository{
		gorm: db,
	}
}
