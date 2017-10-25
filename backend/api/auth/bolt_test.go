package auth_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/api/auth"
)

func TestCreateAdminUser(t *testing.T) {
	// Given empty database
	db, err := bolt.Open("temp.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()
	defer os.Remove("temp.db")
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}

	// When we run create Admin
	manager := auth.NewManager(db)

	// Then a new admin should exist
	_, err = manager.FindByUsername("admin")
	if err != nil {
		t.Fail()
	}
}

func TestDontCreateAdminUser(t *testing.T) {
	// Given an administrator already exists
	db, err := bolt.Open("temp.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()
	defer os.Remove("temp.db")
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}
	manager := auth.NewManager(db)
	admin, _ := manager.FindByUsername("admin")
	manager = auth.NewManager(db)

	// Then nothing should change (compare hashed passwords)
	nextAdmin, _ := manager.FindByUsername("admin")
	assert.Equal(t, admin.HashedPassword, nextAdmin.HashedPassword)
}

func TestUser(t *testing.T) {
	db, err := bolt.Open("temp.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	defer db.Close()
	defer os.Remove("temp.db")
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}
	manager := auth.NewManager(db)

	// Given a user
	u := &auth.User{
		Username: "Bob",
	}
	u.HashPassword("foobar")

	// When the user is saved
	manager.Save(u)

	// Then we should be able to fetch and authenticate with them
	savedUser, err := manager.FindByUsername("Bob")
	assert.Nil(t, err)
	assert.True(t, savedUser.ValidatePassword("foobar"))
}
