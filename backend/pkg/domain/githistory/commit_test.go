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
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
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
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
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

	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	ts := time.Now().UTC()

	c := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts)

	s.NotNil(c)

	s.Equal(p, c.Project)
	s.Equal("abcdef", c.Hash)
	s.Equal("test commit", c.Message)
	s.Equal("me@velocityci.io", c.Author)
	s.Equal(ts, c.CreatedAt)
}

func (s *CommitSuite) TestGetByProjectAndHash() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	b := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(b)

	ts := time.Now()
	nC := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts)

	s.branchManager.SaveCommitToBranch(nC, b)

	c, err := m.GetByProjectAndHash(p, nC.Hash)
	s.Nil(err)
	s.NotNil(c)

	s.EqualValues(nC, c)
}

func (s *CommitSuite) TestGetAllForProject() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	b := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(b)

	ts := time.Now()

	c1 := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts)
	c2 := m.New(p, "123456", "2est commit", "me@velocityci.io", ts)

	s.branchManager.SaveCommitToBranch(c1, b)
	s.branchManager.SaveCommitToBranch(c2, b)

	cs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(cs, 2)
	s.Contains(cs, c1)
	s.Contains(cs, c2)
}

func (s *CommitSuite) TestGetAllForBranch() {
	m := githistory.NewCommitManager(s.storm)

	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	b := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(b)

	ts := time.Now()

	c1 := m.New(p, "abcdef", "test commit", "me@velocityci.io", ts)
	c2 := m.New(p, "123456", "2est commit", "me@velocityci.io", ts)

	s.branchManager.SaveCommitToBranch(c1, b)
	s.branchManager.SaveCommitToBranch(c2, b)

	cs, total := m.GetAllForBranch(b, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(cs, 2)
	s.Contains(cs, c1)
	s.Contains(cs, c2)
}
