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

type CommitSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	branchManager  *githistory.BranchManager
}

func TestCommitSuite(t *testing.T) {
	suite.Run(t, new(CommitSuite))
}

func (s *CommitSuite) SetupTest() {
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
	s.branchManager = githistory.NewBranchManager(s.storm)
}

func (s *CommitSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *CommitSuite) TestNew() {

	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")

	ts := time.Now().UTC()
	c := m.Create(b, p, "abcdef", "test commit", "me@velocityci.io", ts)

	s.NotNil(c)

	s.Equal(p, c.Project)
	s.Equal("abcdef", c.Hash)
	s.Equal("test commit", c.Message)
	s.Equal("me@velocityci.io", c.Author)
	s.Equal(ts, c.CreatedAt)
}

func (s *CommitSuite) TestGetByProjectAndHash() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")

	ts := time.Now()
	nC := m.Create(b, p, "abcdef", "test commit", "me@velocityci.io", ts)

	c, err := m.GetByProjectAndHash(p, nC.Hash)
	s.Nil(err)
	s.NotNil(c)

	s.EqualValues(nC, c)
}

func (s *CommitSuite) TestGetAllForProject() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")

	c1 := m.Create(b, p, "abcdef", "test commit", "me@velocityci.io", time.Now())
	c2 := m.Create(b, p, "123456", "2est commit", "me@velocityci.io", time.Now().Add(1*time.Second))

	cs, total := m.GetAllForProject(p, &githistory.CommitQuery{
		Limit: 5,
		Page:  1,
	})

	s.Equal(2, total)
	s.Len(cs, 2)
	s.Contains(cs, c1)
	s.Contains(cs, c2)

	cs, total = m.GetAllForProject(p, &githistory.CommitQuery{
		Limit: 1,
		Page:  1,
	})

	s.Equal(2, total)
	s.Len(cs, 1)
	s.Contains(cs, c2)

	cs, total = m.GetAllForProject(p, &githistory.CommitQuery{
		Limit: 1,
		Page:  2,
	})

	s.Equal(2, total)
	s.Len(cs, 1)
	s.Contains(cs, c1)
}

func (s *CommitSuite) TestGetAllForProjectBranchFilter() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b1 := s.branchManager.Create(p, "testBranch1")
	b2 := s.branchManager.Create(p, "testBranch2")

	m.Create(b1, p, "abcdef", "test commit", "me@velocityci.io", time.Now())
	c2 := m.Create(b2, p, "123456", "2est commit", "me@velocityci.io", time.Now().Add(1*time.Second))
	c3 := m.Create(b2, p, "1234567", "2est commit", "me@velocityci.io", time.Now().Add(2*time.Second))

	cs, total := m.GetAllForProject(p, &githistory.CommitQuery{
		Limit:    5,
		Page:     1,
		Branches: []string{"testBranch2"},
	})

	s.Equal(2, total)
	s.Len(cs, 2)
	s.Contains(cs, c2)
	s.Contains(cs, c3)

	cs, total = m.GetAllForProject(p, &githistory.CommitQuery{
		Limit:    1,
		Page:     1,
		Branches: []string{"testBranch2"},
	})

	s.Equal(2, total)
	s.Len(cs, 1)
	s.Contains(cs, c3)

	cs, total = m.GetAllForProject(p, &githistory.CommitQuery{
		Limit:    1,
		Page:     2,
		Branches: []string{"testBranch2"},
	})

	s.Equal(2, total)
	s.Len(cs, 1)
	s.Contains(cs, c2)
}

func (s *CommitSuite) TestGetAllForBranch() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")

	ts := time.Now()

	c1 := m.Create(b, p, "abcdef", "test commit", "me@velocityci.io", ts)
	c2 := m.Create(b, p, "123456", "2est commit", "me@velocityci.io", ts)

	cs, total := m.GetAllForBranch(b, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(cs, 2)
	s.Contains(cs, c1)
	s.Contains(cs, c2)
}
