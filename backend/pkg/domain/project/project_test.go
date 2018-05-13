package project_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type ProjectSuite struct {
	suite.Suite
	storm  *storm.DB
	dbPath string
}

func TestProjectSuite(t *testing.T) {
	suite.Run(t, new(ProjectSuite))
}

func (s *ProjectSuite) SetupTest() {
	// Retrieve a temporary path.
	f, err := ioutil.TempFile("", "")
	if err != nil {
		panic(err)
	}
	s.dbPath = f.Name()
	f.Close()
	os.Remove(s.dbPath)
	// Open the database.
	s.storm, err = storm.Open(s.dbPath)
	if err != nil {
		panic(err)
	}
}

func (s *ProjectSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *ProjectSuite) TestValidNew() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, errs := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	s.Nil(errs)

	s.NotEmpty(p.ID)
	s.Equal("Test Project", p.Name)
	s.Equal("testGit", p.Config.Address)
	s.WithinDuration(time.Now().UTC(), p.CreatedAt, 1*time.Second)
	s.WithinDuration(time.Now().UTC(), p.UpdatedAt, 1*time.Second)
}

func (s *ProjectSuite) TestSSHInvalidCreate() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.RawRepository, error) {
		return nil, velocity.SSHKeyError("")
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, errs := m.Create("Test Project", velocity.GitRepository{
		Address:    "testGit",
		PrivateKey: "malformedKey",
	})
	s.Nil(p)
	s.NotNil(errs)

	// s.Equal("", errs.ErrorMap["key"])
	// s.Equal("", errs.ErrorMap["repository"])
}

func (s *ProjectSuite) TestDuplicateCreate() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})
	p, errs := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	s.Nil(p)
	s.NotNil(errs)

	s.Equal([]string{"name already exists!"}, errs.ErrorMap["name"])
}

func (s *ProjectSuite) TestUpdate() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	eP, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	eP.Synchronising = true
	err := m.Update(eP)
	s.Nil(err)

	p, _ := m.GetByName("Test Project")
	s.Equal(eP, p)
}

func (s *ProjectSuite) TestExists() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	s.True(m.Exists("Test Project"))
}

func (s *ProjectSuite) TestDelete() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	err := m.Delete(p)
	s.Nil(err)

	s.False(m.Exists("Test Project"))
}

func (s *ProjectSuite) TestList() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	q := &domain.PagingQuery{
		Limit: 3,
		Page:  1,
	}
	items, amount := m.GetAll(q)
	s.Len(items, 1)
	s.Equal(amount, 1)
	s.Contains(items, p)
}

func (s *ProjectSuite) TestGetByName() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	pR, err := m.GetByName("Test Project")
	s.Nil(err)
	s.Equal(p, pR)
}

func (s *ProjectSuite) TestGetBySlug() {
	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	m := project.NewManager(s.storm, validator, translator, syncMock)

	p, _ := m.Create("Test Project", velocity.GitRepository{
		Address: "testGit",
	})

	pR, err := m.GetBySlug(p.Slug)
	s.Nil(err)
	s.Equal(p, pR)
}
