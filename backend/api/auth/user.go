package auth

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	u.HashedPassword = string(hashedPassword[:])
}

type UserAuth struct {
	Username string    `json:"username"`
	Token    string    `json:"authToken"`
	Expires  time.Time `json:"expires"`
}
