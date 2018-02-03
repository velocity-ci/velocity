package user

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type gormUser struct {
	UUID           string `gorm:"primary_key"`
	Username       string `gorm:"not null"`
	HashedPassword string `gorm:"not null"`
}

func (gormUser) TableName() string {
	return "users"
}

func (gU *gormUser) toUser() *User {
	return &User{
		UUID:           gU.UUID,
		Username:       gU.Username,
		HashedPassword: gU.HashedPassword,
	}
}

func (u *User) toGormUser() *gormUser {
	return &gormUser{
		UUID:           u.UUID,
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
	}
}

type db struct {
	db *gorm.DB
}

func newDB(gorm *gorm.DB) *db {
	return &db{
		db: gorm,
	}
}

func (db *db) save(u *User) error {
	tx := db.db.Begin()

	gU := u.toGormUser()

	tx.
		Where(gormUser{UUID: u.UUID}).
		Assign(&gU).
		FirstOrCreate(&gU)

	return tx.Commit().Error
}

func (db *db) delete(u *User) error {
	tx := db.db.Begin()

	gU := u.toGormUser()

	if err := tx.Delete(gU).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (db *db) getByUsername(username string) (*User, error) {
	gU := gormUser{}
	if db.db.Where("username = ?", username).First(&gU).RecordNotFound() {
		return nil, fmt.Errorf("could not find user %s", username)
	}

	return gU.toUser(), nil
}
