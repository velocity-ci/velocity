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
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
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
	syncMock := func(*velocity.GitRepository, bool, bool, bool, io.Writer) (*git.Repository, string, error) {
		return &git.Repository{}, "/testDir", nil
	}
	s.projectManager = project.NewManager(s.storm, validator, translator, syncMock)
	s.commitManager = githistory.NewCommitManager(s.storm)
	s.branchManager = githistory.NewBranchManager(s.storm)
}

func (s *CommitSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *CommitSuite) TestNew() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())

	m := task.NewManager(s.storm)
	setupStep := velocity.NewSetup()
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	}, setupStep)

	s.NotNil(tsk)

	s.Equal(c, tsk.Commit)
	s.Equal("testTask", tsk.Name)
	s.Equal([]velocity.Step{setupStep}, tsk.Steps)
}

func (s *CommitSuite) TestSave() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())

	m := task.NewManager(s.storm)
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	err := m.Save(tsk)
	s.Nil(err)
}

func (s *CommitSuite) TestGetByCommitAndName() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	b := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(b)

	s.branchManager.SaveCommitToBranch(c, b)

	m := task.NewManager(s.storm)
	tsk := m.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m.Save(tsk)

	rTsk, err := m.GetByCommitAndName(c, "testTask")
	s.NotNil(rTsk)
	s.Nil(err)

	s.Equal(tsk, rTsk)
}

func (s *CommitSuite) TestGetAllForCommit() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	b := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(b)

	s.branchManager.SaveCommitToBranch(c, b)

	m := task.NewManager(s.storm)
	tsk1 := m.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	tsk2 := m.New(c, &velocity.Task{
		Name: "2estTask",
	}, velocity.NewSetup())

	m.Save(tsk1)
	m.Save(tsk2)

	rTsks, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(2, total)
	s.Len(rTsks, 2)

	s.Contains(rTsks, tsk1)
	s.Contains(rTsks, tsk2)
}
