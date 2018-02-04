package user

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
)

type stormUser struct {
	ID             string `storm:"id"`
	Username       string `storm:"index"`
	HashedPassword string
}

func (s *stormUser) ToUser() *User {
	return &User{
		UUID:           s.ID,
		Username:       s.Username,
		HashedPassword: s.HashedPassword,
	}
}

func (u *User) toStormUser() *stormUser {
	return &stormUser{
		ID:             u.UUID,
		Username:       u.Username,
		HashedPassword: u.HashedPassword,
	}
}

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

	if err := tx.Save(u.toStormUser()); err != nil {
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

	tx.DeleteStruct(u.toStormUser())

	return tx.Commit()
}

func (db *stormDB) getByUsername(username string) (*User, error) {
	query := db.Select(q.Eq("Username", username))
	var u stormUser
	if err := query.First(&u); err != nil {
		return nil, err
	}

	return u.ToUser(), nil
}

func GetByUUID(db *storm.DB, uuid string) (*User, error) {
	var u stormUser
	if err := db.One("ID", uuid, &u); err != nil {
		return nil, err
	}
	return u.ToUser(), nil
}
