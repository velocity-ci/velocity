package user

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
	db.AutoMigrate(User{})
	return &gormRepository{
		logger: log.New(os.Stdout, "[gorm:user]", log.Lshortfile),
		gorm:   db,
	}
}

func (r *gormRepository) Save(u User) User {
	tx := r.gorm.Begin()

	tx.
		Where(User{Username: u.Username}).
		Assign(&u).
		FirstOrCreate(&u)

	tx.Commit()
	return u
}

func (r *gormRepository) Delete(u User) {
	tx := r.gorm.Begin()

	if err := tx.Delete(u).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByUsername(username string) (User, error) {
	u := User{}
	if r.gorm.Where(&User{Username: username}).First(&u).RecordNotFound() {
		log.Printf("Could not find User %s", username)
		return User{}, fmt.Errorf("could not find User %s", username)
	}

	return u, nil
}

func (r *gormRepository) GetAll(q Query) ([]User, uint64) {
	users := []User{}
	var count uint64
	r.gorm.Find(&users).Count(count)

	return users, count
}
