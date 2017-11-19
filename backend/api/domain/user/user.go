package user

import (
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	Save(u *User) *User
	Delete(u *User)
	GetByUsername(ID string) (*User, error)
	GetAll(q Query) ([]*User, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
}

type RequestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) ValidatePassword(password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil {
		return true
	}
	return false
}

func (u *User) HashPassword(password string) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	u.HashedPassword = string(hashedPassword[:])
}
