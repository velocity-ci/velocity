package user

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UUID           string `json:"uuid" validate:"-"`
	Username       string `json:"username" validate:"required,min=3"`
	Password       string `json:"password" validate:"required,min=3"`
	HashedPassword string `json:"hashedPassword" validate:"-"`
}

func (u *User) ValidatePassword(password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil {
		return true
	}
	return false
}

func (u *User) hashPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.HashedPassword = string(hashedPassword[:])
	u.Password = ""
}
