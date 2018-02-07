package project

import (
	"io"
	"time"

	"github.com/asdine/storm"
	ut "github.com/go-playground/universal-translator"
	"github.com/gosimple/slug"
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/velocity"
	govalidator "gopkg.in/go-playground/validator.v9"
	git "gopkg.in/src-d/go-git.v4"
)

type Manager struct {
	validator *validator
	db        *stormDB
	clone     func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error)
}

func NewManager(
	db *storm.DB,
	validator *govalidator.Validate,
	translator ut.Translator,
	cloneFunc func(r *velocity.GitRepository, bare bool, full bool, submodule bool, writer io.Writer) (*git.Repository, string, error),
) *Manager {
	m := &Manager{
		db:    newStormDB(db),
		clone: cloneFunc,
	}
	m.validator = newValidator(validator, translator, m)
	return m
}

func (m *Manager) Create(name string, config velocity.GitRepository) (*Project, *domain.ValidationErrors) {
	p := &Project{
		Name:      name,
		Config:    config,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := m.validator.Validate(p); err != nil {
		return nil, err
	}

	p.ID = uuid.NewV1().String()
	p.Slug = slug.Make(p.Name)

	m.db.save(p)

	return p, nil
}

func (m *Manager) Update(p *Project) error {
	return m.db.save(p)
}

func (m *Manager) Delete(p *Project) error {
	return m.db.delete(p)
}

func (m *Manager) Exists(name string) bool {
	if _, err := m.GetBySlug(slug.Make(name)); err != nil {
		return false
	}
	return true
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
