package user

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type stormDB struct {
	*storm.DB
}

func newStormDB(db *storm.DB) *stormDB {
	db.Init(&User{})
	return &stormDB{db}
}

func (db *stormDB) save(u *User) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	if err := tx.Save(u); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *stormDB) delete(u *User) error {
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}

	tx.DeleteStruct(u)

	return tx.Commit()
}

func (db *stormDB) getByUsername(username string) (*User, error) {
	query := db.Select(q.Eq("Username", username))
	var u User
	if err := query.First(&u); err != nil {
		return nil, err
	}

	return &u, nil
}

func GetByUUID(db *storm.DB, uuid string) (*User, error) {
	var u User
	if err := db.One("UUID", uuid, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
