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

type StreamSuite struct {
	suite.Suite
	storm          *storm.DB
	dbPath         string
	projectManager *project.Manager
	commitManager  *githistory.CommitManager
	branchManager  *githistory.BranchManager
	taskManager    *task.Manager
	buildManager   *build.BuildManager
	stepManager    *build.StepManager
	streamManager  *build.StreamManager
	wg             sync.WaitGroup
}

func TestStreamSuite(t *testing.T) {
	suite.Run(t, new(StreamSuite))
}

func (s *StreamSuite) SetupTest() {
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
	s.buildManager = build.NewBuildManager(s.storm, s.stepManager, s.streamManager)
}

func (s *StreamSuite) TearDownTest() {
	defer os.Remove(s.dbPath)
	s.wg.Wait()
	s.storm.Close()
}

func (s *StreamSuite) TestFileStreamLine() {
	p, _ := s.projectManager.Create("testProject", velocity.GitRepository{
		Address: "testGit",
	})

	br := s.branchManager.Create(p, "testBranch")
	c := s.commitManager.Create(br, p, "abcdef", "test commit", "me@velocityci.io", time.Now().UTC(), "")

	tsk := s.taskManager.Create(c, &velocity.Task{
		Name: "testTask",
	}, velocity.NewSetup())

	params := map[string]string{}
	b, _ := s.buildManager.Create(tsk, params)

	steps := s.stepManager.GetStepsForBuild(b)

	streams := s.streamManager.GetStreamsForStep(steps[0])

	stream := streams[0]

	timestamp := time.Now().UTC()
	streamLine := s.streamManager.CreateStreamLine(stream, 1, timestamp, "test output")
	s.Equal(stream.ID, streamLine.StreamID)
	s.Equal(1, streamLine.LineNumber)
	s.Equal(timestamp, streamLine.Timestamp)
	s.Equal("test output\n", streamLine.Output)

	streamLines, total := s.streamManager.GetStreamLines(stream, domain.NewPagingQuery())

	s.Equal(1, total)
	s.Len(streamLines, 1)
	s.Contains(streamLines, streamLine)
}
