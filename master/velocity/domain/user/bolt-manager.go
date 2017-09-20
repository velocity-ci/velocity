package user

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/utils"
)

type BoltManager struct {
	logger *log.Logger
	bolt   *bolt.DB
}

func createAdminUser(m *BoltManager) {
	_, err := m.FindByUsername("admin")
	if err != nil {
		password := utils.GenerateRandomString(16)
		user := &domain.BoltUser{Username: "admin"}
		user.HashPassword(password)
		m.Save(user)
		m.logger.Printf("\n\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n\n", password)
	}
}

func NewBoltManager(
	boltLogger *log.Logger,
	bolt *bolt.DB,
) *BoltManager {
	m := &BoltManager{
		logger: boltLogger,
		bolt:   bolt,
	}
	createAdminUser(m)
	return m
}

func (m *BoltManager) Save(u *domain.BoltUser) error {
	tx, err := m.bolt.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	m.logger.Printf("saving user: %s", u.Username)

	usersBucket, err := tx.CreateBucketIfNotExists([]byte("users"))
	if usersBucket == nil {
		usersBucket = tx.Bucket([]byte("users"))
	}

	userJSON, err := json.Marshal(u)
	if err != nil {
		return err
	}
	usersBucket.Put([]byte(u.Username), userJSON)

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *BoltManager) FindByUsername(username string) (*domain.BoltUser, error) {
	tx, err := m.bolt.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	usersBucket := tx.Bucket([]byte("users"))
	if usersBucket == nil {
		return nil, fmt.Errorf("User not found: %s", username)
	}

	v := usersBucket.Get([]byte(username))

	u := domain.BoltUser{}
	err = json.Unmarshal(v, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
