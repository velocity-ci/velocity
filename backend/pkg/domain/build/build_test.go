package build_test

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/asdine/storm"
	"github.com/stretchr/testify/suite"
	"github.com/velocity-ci/velocity/backend/pkg/domain"
	"github.com/velocity-ci/velocity/backend/pkg/domain/build"
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

type BuildSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	branchManager  *githistory.BranchManager
	taskManager    *task.Manager
	stepManager    *build.StepManager
	streamManager  *build.StreamManager
	wg             sync.WaitGroup
}

var syncMock = func(*velocity.GitRepository) (bool, error) {
	return true, nil
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
	s.projectManager = project.NewManager(s.storm, validator, translator, syncMock)
	s.commitManager = githistory.NewCommitManager(s.storm)
	s.branchManager = githistory.NewBranchManager(s.storm)
	s.taskManager = task.NewManager(s.storm, s.projectManager, s.branchManager, s.commitManager)
	s.stepManager = build.NewStepManager(s.storm)
	s.streamManager = build.NewStreamManager(s.storm)
}

func (s *BuildSuite) TearDownTest() {
	s.wg.Wait()
	defer os.Remove(s.dbPath)
	s.storm.Close()
}

func (s *BuildSuite) TestNewBuild() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	s.Equal(tsk, b.Task)
	s.Equal(params, b.Parameters)
	s.Equal(velocity.StateWaiting, b.Status)
	s.WithinDuration(time.Now().UTC(), b.CreatedAt, 1*time.Second)
	s.WithinDuration(time.Now().UTC(), b.UpdatedAt, 1*time.Second)

	steps := s.stepManager.GetStepsForBuild(b)
	s.Len(steps, len(tsk.VTask.Steps))
}

func (s *BuildSuite) TestUpdateBuild() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	err := m.Update(b)
	s.Nil(err)
}

func (s *BuildSuite) TestGetBuildsForProject() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForProject(p, &build.BuildQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForProjectFilter() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)
	b.Status = "running"
	err := m.Update(b)
	s.Nil(err)
	_, errs = m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForProject(p, &build.BuildQuery{Limit: 5, Page: 1, Status: "running"})

	s.Equal(1, total)
	s.Len(rbs, 1)
	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForCommit() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForCommit(c, &build.BuildQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForCommitFilter() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)
	b.Status = "running"
	err := m.Update(b)
	s.Nil(err)
	_, errs = m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForCommit(c, &build.BuildQuery{Limit: 5, Page: 1, Status: "running"})

	s.Equal(1, total)
	s.Len(rbs, 1)
	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForTask() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForTask(tsk, &build.BuildQuery{Limit: 5, Page: 1})

	s.Equal(1, total)
	s.Len(rbs, 1)

	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildsForTaskFilter() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)
	b.Status = "running"
	err := m.Update(b)
	s.Nil(err)
	_, errs = m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetAllForTask(tsk, &build.BuildQuery{Limit: 5, Page: 1, Status: "running"})

	s.Equal(1, total)
	s.Len(rbs, 1)
	s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetRunningBuilds() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)
	b.Status = velocity.StateRunning
	m.Update(b)

	rbs, total := m.GetRunningBuilds()

	s.Equal(1, total)
	s.Len(rbs, 1)

	// s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetWaitingBuilds() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	_, errs := m.Create(tsk, params)
	s.Nil(errs)

	rbs, total := m.GetWaitingBuilds()

	s.Equal(1, total)
	s.Len(rbs, 1)

	// s.Equal(b, rbs[0])
}

func (s *BuildSuite) TestGetBuildByID() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	m := build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
	params := map[string]string{}
	b, errs := m.Create(tsk, params)
	s.Nil(errs)

	rB, err := m.GetBuildByID(b.ID)
	s.Nil(err)
	s.Equal(b, rB)
}
