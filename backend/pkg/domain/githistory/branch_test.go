package githistory_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type BranchSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
}

func TestBranchSuite(t *testing.T) {
	suite.Run(t, new(BranchSuite))
}

func (s *BranchSuite) SetupTest() {
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

	validator, translator := domain.NewValidator()
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*velocity.RawRepository, error) {
		return &velocity.RawRepository{Directory: "/testDir"}, nil
	}
	s.projectManager = project.NewManager(s.storm, validator, translator, syncMock)
	s.commitManager = githistory.NewCommitManager(s.storm)
}

func (s *BranchSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *BranchSuite) TestCreate() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b := m.Create(p, "testBranch")
	s.NotNil(b)

	s.Equal(p, b.Project)
	s.Equal("testBranch", b.Name)
}

func (s *BranchSuite) TestUpdate() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b := m.Create(p, "testBranch")

	b.Active = false
	err := m.Update(b)
	s.Nil(err)

	aB, err := m.GetByProjectAndName(p, "testBranch")
	s.Nil(err)
	s.NotNil(aB)
	s.Equal(b, aB)
}

func (s *BranchSuite) TestGetByProjectAndName() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	m.Create(p, "testBranch")

	aB, err := m.GetByProjectAndName(p, "testBranch")
	s.Nil(err)
	s.NotNil(aB)
}

func (s *BranchSuite) TestGetAllForProject() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b1 := m.Create(p, "testBranch")
	b2 := m.Create(p, "2estBranch")

	bs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(bs, 2)
	s.Contains(bs, b1)
	s.Contains(bs, b2)
}

func (s *BranchSuite) TestGetAllForCommit() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b1 := m.Create(p, "testBranch")
	b2 := m.Create(p, "2estBranch")

	c := s.commitManager.Create(b1, p, "abcdef", "test commit", "me@velocityci.io", time.Now())
	s.commitManager.AddCommitToBranch(c, b2)

	bs, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(bs, 2)
	s.Contains(bs, b1)
	s.Contains(bs, b2)
}
