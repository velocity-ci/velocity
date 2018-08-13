package user

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             string `json:"id" validate:"-" storm:"id"`
	Username       string `json:"username" validate:"required,min=3,alphanum,userUnique"`
	Password       string `json:"password" validate:"required,min=3"`
	HashedPassword string `json:"hashedPassword" validate:"-"`
}

func (u *User) ValidatePassword(password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

func (u *User) hashPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.HashedPassword = string(hashedPassword[:])
	u.Password = ""
}
