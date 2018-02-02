package domain

import (
	"github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	validator "gopkg.in/go-playground/validator.v9"
)

type User struct {
	UUID           string `json:"uuid" validate:"-"`
	Username       string `json:"username" validate:"required,min=3"`
	Password       string `json:"-" validate:"required,min=3"`
	HashedPassword string `json:"hashedPassword" validate:"-"`
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

func (u *User) RegisterCustomValidations(v *validator.Validate, t ut.Translator) {}

func NewUser(username string, password string) (*User, validator.ValidationErrors) {
	u := &User{
		UUID:     uuid.NewV1().String(),
		Username: username,
		Password: password,
	}

	err := validate.Struct(u)
	if err != nil {
		return nil, err.(validator.ValidationErrors)
	}
	u.HashPassword(password)

	return u, nil
}
