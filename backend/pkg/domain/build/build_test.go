package build_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
	git "gopkg.in/src-d/go-git.v4"
)

type BuildSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	branchManager  *githistory.BranchManager
	taskManager    *task.Manager
}

func TestBuildSuite(t *testing.T) {
	suite.Run(t, new(BuildSuite))
}

func (s *BuildSuite) SetupTest() {
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
	s.taskManager = task.NewManager(s.storm)
}

func (s *BuildSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *BuildSuite) TestNewBuild() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)

	s.Equal(tsk, b.Task)
	s.Equal(params, b.Parameters)
	s.Equal(velocity.StateWaiting, b.Status)
	s.WithinDuration(time.Now().UTC(), b.CreatedAt, 1*time.Second)
	s.WithinDuration(time.Now().UTC(), b.UpdatedAt, 1*time.Second)

	s.Len(b.Steps, len(tsk.Steps))
}

func (s *BuildSuite) TestSaveBuild() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)

	err := m.Save(b)
	s.Nil(err)
}

func (s *BuildSuite) TestGetBuildsForProject() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForProject(p, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForCommit() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForCommit(c, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForTask() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetAllForTask(tsk, &domain.PagingQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetRunningBuilds() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	b.Status = velocity.StateRunning
	m.Save(b)

	rbs, total := m.GetRunningBuilds()

	s.Equal(1, total)
	s.Len(rbs, 1)

	// s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetWaitingBuilds() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rbs, total := m.GetWaitingBuilds()

	s.Equal(1, total)
	s.Len(rbs, 1)

	// s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildByUUID() {
	p, _ := s.projectManager.New("testProject", velocity.GitRepository{
		Address: "testGit",
	})
	s.projectManager.Save(p)

	c := s.commitManager.New(p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC())
	br := s.branchManager.New(p, "testBranch")
	s.branchManager.Save(br)

	s.branchManager.SaveCommitToBranch(c, br)

	tsk := s.taskManager.New(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())
	s.taskManager.Save(tsk)

	m := build.NewBuildManager(s.storm, build.NewStepManager(s.storm), build.NewStreamManager(s.storm))
	params := map[string]string{}
	b := m.New(tsk, params)
	m.Save(b)

	rB, err := m.GetBuildByUUID(b.UUID)
	s.Nil(err)
	s.Equal(b, rB)
}
