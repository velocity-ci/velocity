package task_test

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
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type CommitSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	branchManager  *githistory.BranchManager
	commitManager  *githistory.CommitManager
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
	s.commitManager = githistory.NewCommitManager(s.storm)
	s.branchManager = githistory.NewBranchManager(s.storm)
}

func (s *CommitSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *CommitSuite) TestCreate() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testProject")

	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())

	m := task.NewManager(s.storm, s.projectManager, s.branchManager, s.commitManager)
	setupStep := velocity.NewSetup()
	tsk := m.Create(c, &velocity.Task{
		Name: "testTask",
	}, setupStep)

	s.NotNil(tsk)

	s.Equal(c, tsk.Commit)
	s.Equal("testtask", tsk.Slug)
	s.Equal([]velocity.Step{setupStep}, tsk.VTask.Steps)
}

func (s *CommitSuite) TestGetByCommitAndSlug() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(b, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())

	m := task.NewManager(s.storm, s.projectManager, s.branchManager, s.commitManager)
	tsk := m.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	rTsk, err := m.GetByCommitAndSlug(c, "testtask")
	s.NotNil(rTsk)
	s.Nil(err)

	s.Equal(tsk, rTsk)
}

func (s *CommitSuite) TestGetAllForCommit() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	b := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(b, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())

	m := task.NewManager(s.storm, s.projectManager, s.branchManager, s.commitManager)
	tsk1 := m.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	tsk2 := m.Create(c, &velocity.Task{
		Name: "2estTask",
	}, velocity.NewSetup())

	rTsks, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(rTsks, 2)

	s.Contains(rTsks, tsk1)
	s.Contains(rTsks, tsk2)
}
