package user_test

import (
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/user"
	govalidator "gopkg.in/go-playground/validator.v9"
)

func setup() (*gorm.DB, *govalidator.Validate, ut.Translator) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()

	return db, validator, translator
}

func TestValidNew(t *testing.T) {
	m := user.NewManager(setup())

	u, errs := m.New("admin", "password")
	assert.Nil(t, errs)

	assert.NotEmpty(t, u.UUID)
	assert.Equal(t, "admin", u.Username)
	assert.Empty(t, u.Password)
	assert.NotEmpty(t, u.HashedPassword)
	assert.True(t, u.ValidatePassword("password"))
}

func TestInvalidNew(t *testing.T) {
	m := user.NewManager(setup())

	u, errs := m.New("ad", "password")
	assert.Nil(t, u)
	assert.NotNil(t, errs)

	assert.Equal(t, []string{"username must be at least 3 characters in length"}, errs.ErrorMap["username"])

	u, errs = m.New("admin", "pa")
	assert.Nil(t, u)
	assert.NotNil(t, errs)

	assert.Equal(t, []string{"password must be at least 3 characters in length"}, errs.ErrorMap["password"])
}

func TestSave(t *testing.T) {
	m := user.NewManager(setup())

	u, _ := m.New("admin", "password")

	err := m.Save(u)

	assert.Nil(t, err)

	assert.True(t, m.Exists("admin"))
}

func TestDelete(t *testing.T) {
	m := user.NewManager(setup())

	u, _ := m.New("admin", "password")
	m.Save(u)

	err := m.Delete(u)
	assert.Nil(t, err)

	assert.False(t, m.Exists("admin"))
}
