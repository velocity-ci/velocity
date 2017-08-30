package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Username       string    `gorm:"primary_key" json:"username"`
	Password       string    `gorm:"-" json:"-"`
	HashedPassword string    `json:"-"`
}

func (u *User) ValidatePassword(password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password)) == nil {
		return true
	}
	return false
}

func (u *User) HashPassword() {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	u.HashedPassword = string(hashedPassword[:])
	u.Password = ""
}

type UserAuth struct {
	Username string    `json:"username"`
	Token    string    `json:"authToken"`
	Expires  time.Time `json:"expires"`
}
