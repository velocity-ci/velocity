package user

import (
	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"go.uber.org/zap"
)

type StormUser struct {
	ID             string `storm:"id"`
	Username       string `storm:"index"`
	HashedPassword string
}

func (s *StormUser) ToUser() *User {
	return &User{
		ID:             s.ID,
		Username:       s.Username,
		HashedPassword: s.HashedPassword,
	}
}

func (u *User) toStormUser() *StormUser {
	return &StormUser{
		ID:             u.ID,
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
	var u StormUser
	if err := query.First(&u); err != nil {
		return nil, err
	}

	return u.ToUser(), nil
}

func GetByID(db *storm.DB, id string) (*User, error) {
	var u StormUser
	if err := db.One("ID", id, &u); err != nil {
		return nil, err
	}
	return u.ToUser(), nil
}

func (db *stormDB) getAll(pQ *domain.PagingQuery) (r []*User, t int) {
	t = 0
	t, err := db.Count(&StormUser{})
	if err != nil {
		velocity.GetLogger().Error("error", zap.Error(err))
		return r, t
	}

	query := db.Select()
	query.Limit(pQ.Limit).Skip((pQ.Page - 1) * pQ.Limit)
	var stormUsers []*StormUser
	query.Find(&stormUsers)

	for _, u := range stormUsers {
		r = append(r, u.ToUser())
	}

	return r, t
}
