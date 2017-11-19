package user

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
)

type GORMUser struct {
	Username       string `gorm:"primary_key"`
	HashedPassword string
}

func gormUserFromUser(u *User) *GORMUser {
	return &GORMUser{
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
	}
}

func userFromGORMUser(g *GORMUser) *User {
	return &User{
		Username:       g.Username,
		HashedPassword: g.HashedPassword,
	}
}

// Expose CRUD operations (implement interface?) Implement repository funcs, as they will be used when we have caching.
type gormRepository struct {
	gorm *gorm.DB
}

func newGORMRepository(db *gorm.DB) *gormRepository {
	db.AutoMigrate(GORMUser{})
	return &gormRepository{
		gorm: db,
	}
}

func (r *gormRepository) Save(u *User) *User {
	tx := r.gorm.Begin()

	gormUser := gormUserFromUser(u)

	tx.
		Where(GORMUser{Username: gormUser.Username}).
		Assign(gormUser).
		FirstOrCreate(gormUser)

	tx.Commit()
	return u
}

func (r *gormRepository) Delete(u *User) {
	tx := r.gorm.Begin()

	gormUser := gormUserFromUser(u)

	if err := tx.Delete(gormUser).Error; err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	tx.Commit()
}

func (r *gormRepository) GetByUsername(username string) (*User, error) {
	gormUser := &GORMUser{}
	if r.gorm.Where(&GORMUser{Username: username}).First(gormUser).RecordNotFound() {
		log.Printf("Could not find User %s", username)
		return nil, fmt.Errorf("could not find User %s", username)
	}

	return userFromGORMUser(gormUser), nil
}

func (r *gormRepository) GetAll(q Query) ([]*User, uint64) {
	gormUsers := []*GORMUser{}
	var count uint64
	r.gorm.Find(gormUsers).Count(count)

	users := []*User{}
	for _, gUser := range gormUsers {
		users = append(users, userFromGORMUser(gUser))
	}

	return users, count
}
