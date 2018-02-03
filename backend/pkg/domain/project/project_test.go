package project_test

import (
	"io"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
	govalidator "gopkg.in/go-playground/validator.v9"
	git "gopkg.in/src-d/go-git.v4"
)

func setup(f func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error)) (*gorm.DB, *govalidator.Validate, ut.Translator, func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error)) {
	db := domain.NewGORMDB(":memory:")
	validator, translator := domain.NewValidator()

	return db, validator, translator, f
}

func TestValidNew(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, errs := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	assert.Nil(t, errs)

	assert.NotEmpty(t, p.UUID)
	assert.Equal(t, "Test Project", p.Name)
	assert.Equal(t, "testGit", p.Config.Address)
}

func TestSSHInvalidNew(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return nil, "", velocity.SSHKeyError("")
	}
	m := project.NewManager(setup(syncMock))

	p, errs := m.New("Test Project", velocity.GitRepository{
		Address:    "testGit",
		PrivateKey: "malformedKey",
	})
	assert.Nil(t, p)
	assert.NotNil(t, errs)

	// assert.Equal(t, "", errs.ErrorMap["key"])
	// assert.Equal(t, "", errs.ErrorMap["repository"])
}

func TestDuplicateNew(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)
	p, errs := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	assert.Nil(t, p)
	assert.NotNil(t, errs)

	assert.Equal(t, []string{"name already exists!"}, errs.ErrorMap["name"])
}

func TestSave(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	err := m.Save(p)
	assert.Nil(t, err)
}

func TestExists(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)

	assert.True(t, m.Exists("Test Project"))
}

func TestDelete(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)

	err := m.Delete(p)
	assert.Nil(t, err)

	assert.False(t, m.Exists("Test Project"))
}

func TestList(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)

	q := &domain.PagingQuery{
		Limit: 3,
		Page:  1,
	}
	items, amount := m.GetAll(q)
	assert.Len(t, items, 1)
	assert.Equal(t, amount, 1)
}

func TestGetByName(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)

	pR, err := m.GetByName("Test Project")
	assert.Nil(t, err)
	assert.Equal(t, p, pR)
}

func TestGetBySlug(t *testing.T) {
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	m := project.NewManager(setup(syncMock))

	p, _ := m.New("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	m.Save(p)

	pR, err := m.GetBySlug(p.Slug)
	assert.Nil(t, err)
	assert.Equal(t, p, pR)
}
