package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

// Manager - Manages users to authenticate against
type Manager struct {
	logger *log.Logger
	bolt   *bolt.DB
}

func createAdminUser(m *Manager) {
	_, err := m.FindByUsername("admin")
	if err != nil {
		var password string
		if os.Getenv("ADMIN_PASSWORD") != "" {
			password = os.Getenv("ADMIN_PASSWORD")
		} else {
			password = GenerateRandomString(16)
		}
		user := &User{Username: "admin"}
		user.HashPassword(password)
		m.Save(user)
		m.logger.Printf("\n\n\nCreated Administrator:\n\tusername: admin \n\tpassword: %s\n\n\n", password)
	}
}

// NewManager - Returns a new auth manager
func NewManager(
	bolt *bolt.DB,
) *Manager {
	m := &Manager{
		logger: log.New(os.Stdout, "[bolt:user]", log.Lshortfile),
		bolt:   bolt,
	}
	createAdminUser(m)
	return m
}

// Save - Saves the given User to persistence
func (m *Manager) Save(u *User) error {
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

	return tx.Commit()
}

// FindByUsername - Finds a User given by their username, returns an error if not found.
func (m *Manager) FindByUsername(username string) (*User, error) {
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

	u := User{}
	err = json.Unmarshal(v, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
