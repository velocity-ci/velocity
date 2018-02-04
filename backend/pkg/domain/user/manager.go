package user

import (
	"github.com/asdine/storm"
	ut "github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	govalidator "gopkg.in/go-playground/validator.v9"
)

type Manager struct {
	validator *validator
	db        *stormDB
}

func NewManager(
	db *storm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
) *Manager {
	m := &Manager{
		db: newStormDB(db),
	}
	m.validator = newValidator(validator, translator, m)

	return m
}

func (m *Manager) New(username, password string) (*User, *domain.ValidationErrors) {
	u := &User{
		Username: username,
		Password: password,
	}

	if err := m.validator.Validate(u); err != nil {
		return nil, err
	}

	u.UUID = uuid.NewV1().String()
	u.hashPassword(password)

	return u, nil
}

func (m *Manager) Save(u *User) error {
	return m.db.save(u)
}

func (m *Manager) Delete(u *User) error {
	return m.db.delete(u)
}

func (m *Manager) Exists(username string) bool {
	if _, err := m.GetByUsername(username); err != nil {
		return false
	}
	return true
}

func (m *Manager) GetByUsername(username string) (*User, error) {
	return m.db.getByUsername(username)
}
