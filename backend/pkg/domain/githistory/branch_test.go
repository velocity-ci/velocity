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
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	s.projectManager = project.NewManager(s.storm, validator, translator, syncMock)
	s.commitManager = githistory.NewCommitManager(s.storm)
}

func (s *BranchSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *BranchSuite) TestNew() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b := m.New(p, "testBranch")
	s.NotNil(b)

	s.Equal(p, b.Project)
	s.Equal("testBranch", b.Name)
}

func (s *BranchSuite) TestSave() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b := m.New(p, "testBranch")

	err := m.Save(b)

	s.Nil(err)
}

func (s *BranchSuite) TestGetAllForProject() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b1 := m.New(p, "testBranch")
	b2 := m.New(p, "2estBranch")

	m.Save(b1)
	m.Save(b2)

	bs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(bs, 2)
	s.Contains(bs, b1)
	s.Contains(bs, b2)
}

func (s *BranchSuite) TestGetAllForCommit() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	m := githistory.NewBranchManager(s.storm)

	b1 := m.New(p, "testBranch")
	b2 := m.New(p, "2estBranch")

	m.Save(b1)
	m.Save(b2)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now())
	m.SaveCommitToBranch(c, b1)
	m.SaveCommitToBranch(c, b2)

	bs, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(bs, 2)
	s.Contains(bs, b1)
	s.Contains(bs, b2)
}
