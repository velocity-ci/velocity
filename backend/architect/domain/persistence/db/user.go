package db

import (
	"fmt"

	"github.com/velocity-ci/velocity/backend/architect/domain"
)

type user struct {
	UUID           string `gorm:"primary_key"`
	Username       string `gorm:"not null;unique"`
	HashedPassword string `gorm:"not null"`
}

func (u user) toDomainUser() *domain.User {
	return &domain.User{
		UUID:           u.UUID,
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
	}
}

func fromDomainUser(u *domain.User) user {
	return user{
		UUID:           u.UUID,
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
	}
}

func SaveUser(u *domain.User) error {
	tx := db.Begin()

	gU := fromDomainUser(u)

	tx.
		Where(user{Username: u.Username}).
		Assign(&gU).
		FirstOrCreate(&gU)

	return tx.Commit().Error
}

func DeleteUser(u *domain.User) error {
	tx := db.Begin()

	gU := fromDomainUser(u)

	if err := tx.Delete(gU).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func GetUserByUsername(username string) (*domain.User, error) {
	gU := user{}
	if db.Where("username = ?", username).First(&gU).RecordNotFound() {
		return nil, fmt.Errorf("could not find user %s", username)
	}

	return gU.toDomainUser(), nil
}
