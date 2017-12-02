package user

import (
	"github.com/jinzhu/gorm"
)

type Manager struct {
	gormRepository *gormRepository
}

func NewManager(
	db *gorm.DB,
) *Manager {
	m := &Manager{
		gormRepository: newGORMRepository(db),
	}
	return m
}

func (m *Manager) Save(u User) User {
	m.gormRepository.Save(u)
	return u
}
func (m *Manager) Delete(u User) {
	m.gormRepository.Delete(u)
}
func (m *Manager) GetByUsername(username string) (User, error) {
	return m.gormRepository.GetByUsername(username)
}
func (m *Manager) GetAll(q Query) ([]User, uint64) {
	return m.gormRepository.GetAll(q)
}
