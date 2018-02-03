package project

import (
	"io"

	ut "github.com/go-playground/universal-translator"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/velocity"
	govalidator "gopkg.in/go-playground/validator.v9"
	git "gopkg.in/src-d/go-git.v4"
)

type Manager struct {
	validator *validator
	db        *db
	Sync      func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error)
}

func NewManager(
	db *gorm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
	syncFunc func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error),
) *Manager {
	db.AutoMigrate(&GormProject{})
	m := &Manager{
		db:   newDB(db),
		Sync: syncFunc,
	}
	m.validator = newValidator(validator, translator, m)
	return m
}

func (m *Manager) New(name string, config velocity.GitRepository) (*Project, *domain.ValidationErrors) {
	p := &Project{
		Name:   name,
		Config: config,
	}

	if err := m.validator.Validate(p); err != nil {
		return nil, err
	}

	p.UUID = uuid.NewV1().String()
	p.Slug = slug.Make(p.Name)

	return p, nil
}

func (m *Manager) Exists(name string) bool {
	if _, err := m.GetBySlug(slug.Make(name)); err != nil {
		return false
	}
	return true
}

func (m *Manager) Save(p *Project) error {
	return m.db.save(p)
}

func (m *Manager) Delete(p *Project) error {
	return m.db.delete(p)
}

func (m *Manager) GetAll(q *domain.PagingQuery) ([]*Project, int) {
	return m.db.getAll(q)
}

func (m *Manager) GetByName(name string) (*Project, error) {
	return m.db.getByName(name)
}

func (m *Manager) GetBySlug(slug string) (*Project, error) {
	return m.db.getBySlug(slug)
}
