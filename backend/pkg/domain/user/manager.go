package user

import (
	"fmt"
	"log"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	ut "github.com/go-playground/universal-translator"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	govalidator "gopkg.in/go-playground/validator.v9"
)

// Event constants
const (
	EventCreate = "user:new"
	EventUpdate = "user:update"
	EventDelete = "user:delete"
)

type Manager struct {
	validator *validator
	db        *stormDB
	brokers   []domain.Broker
}

func NewManager(
	db *storm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
) *Manager {
	m := &Manager{
		db:      newStormDB(db),
		brokers: []domain.Broker{},
	}
	m.validator = newValidator(validator, translator, m)

	return m
}

func (m *Manager) AddBroker(b domain.Broker) {
	m.brokers = append(m.brokers, b)
}

func (m *Manager) EnsureAdminUser() {

	if !m.Exists("admin") {
		var password string
		if os.Getenv("ADMIN_PASSWORD") != "" {
			password = os.Getenv("ADMIN_PASSWORD")
		} else {
			password = GenerateRandomString(16)
		}
		_, err := m.Create("admin", password)
		if err != nil {
			logrus.Error(err)
		}
		log.Printf("\n\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n\n", password)
	}
}

func (m *Manager) Create(username, password string) (*User, *domain.ValidationErrors) {
	u := &User{
		Username: username,
		Password: password,
	}

	if err := m.validator.Validate(u); err != nil {
		return nil, err
	}

	u.ID = uuid.NewV1().String()
	u.hashPassword(password)

	if err := m.db.save(u); err != nil {
		logrus.Error(err)
		return nil, nil
	}

	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   "users",
			Event:   EventCreate,
			Payload: u,
		})
	}

	return u, nil
}

func (m *Manager) Update(u *User) error {
	if err := m.db.save(u); err != nil {
		return err
	}
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   fmt.Sprintf("user:%s", u.ID),
			Event:   EventUpdate,
			Payload: u,
		})
	}
	return nil
}

func (m *Manager) Delete(u *User) error {
	if err := m.db.delete(u); err != nil {
		return err
	}
	for _, b := range m.brokers {
		b.EmitAll(&domain.Emit{
			Topic:   fmt.Sprintf("user:%s", u.ID),
			Event:   EventDelete,
			Payload: u,
		})
	}
	return nil
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

func (m *Manager) GetAll(q *domain.PagingQuery) ([]*User, int) {
	return m.db.getAll(q)
}
