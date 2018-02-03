package project

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type GormProject struct {
	UUID   string `gorm:"primary_key"`
	Slug   string `gorm:"not null"`
	Name   string `gorm:"not null"`
	Config []byte
}

func (GormProject) TableName() string {
	return "projects"
}

func (gP *GormProject) ToProject() *Project {
	config := velocity.GitRepository{}
	err := json.Unmarshal(gP.Config, &config)
	if err != nil {
		logrus.Error(err)
	}
	return &Project{
		UUID:   gP.UUID,
		Name:   gP.Name,
		Slug:   gP.Slug,
		Config: config,
	}
}

func (p *Project) ToGormProject() *GormProject {
	jsonConfig, err := json.Marshal(p.Config)
	if err != nil {
		logrus.Error(err)
	}

	return &GormProject{
		UUID:   p.UUID,
		Name:   p.Name,
		Slug:   p.Slug,
		Config: jsonConfig,
	}
}

type db struct {
	db *gorm.DB
}

func newDB(gorm *gorm.DB) *db {
	return &db{
		db: gorm,
	}
}

func (db *db) delete(p *Project) error {
	tx := db.db.Begin()

	gP := p.ToGormProject()

	if err := tx.Delete(gP).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (db *db) save(p *Project) error {
	tx := db.db.Begin()

	gP := p.ToGormProject()

	tx.
		Where(GormProject{UUID: p.UUID}).
		Assign(&gP).
		FirstOrCreate(&gP)

	return tx.Commit().Error
}

func (db *db) getAll(q *domain.PagingQuery) (r []*Project, t int) {
	t = 0

	gPs := []GormProject{}
	db.db.Find(&gPs).Count(&t)

	db.db.
		Limit(q.Limit).
		Offset((q.Page - 1) * q.Limit).
		Find(&gPs)

	for _, gP := range gPs {
		r = append(r, gP.ToProject())
	}

	return r, t
}

func (db *db) getByName(name string) (*Project, error) {
	gP := GormProject{}
	if db.db.Where("name = ?", name).First(&gP).RecordNotFound() {
		return nil, fmt.Errorf("could not find project %s", name)
	}

	return gP.ToProject(), nil
}

func (db *db) getBySlug(slug string) (*Project, error) {
	gP := GormProject{}
	if db.db.Where("slug = ?", slug).First(&gP).RecordNotFound() {
		return nil, fmt.Errorf("could not find project %s", slug)
	}

	return gP.ToProject(), nil
}
