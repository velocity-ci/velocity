package build

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	SaveBuild(b Build) Build
	DeleteBuild(b Build)
	GetBuildByProjectAndCommitAndID(p project.Project, c commit.Commit, id string) (Build, error)
	// Order timestamp descending
	GetBuildsByProject(p project.Project, q Query) ([]Build, uint64)
	// Order timestamp descending
	GetBuildsByProjectAndCommit(p project.Project, c commit.Commit) ([]Build, uint64)
	GetRunningBuilds() ([]Build, uint64)
	GetWaitingBuilds() ([]Build, uint64)

	// BuildSteps
	SaveBuildStep(bS BuildStep) BuildStep
	GetBuildStepsForBuild(b Build) ([]BuildStep, uint64)
	GetBuildStepByBuildAndID(b Build, id string) (BuildStep, error)

	// OutputStreams
	SaveOutputStream(oS velocity.OutputStream) velocity.OutputStream
	GetOutputStreamsForBuildStep(bS BuildStep) ([]velocity.OutputStream, uint64)
	GetOutputStreamByID(id string) (velocity.OutputStream, error)

	// StreamLines
	SaveStreamLine(sL StreamLine) StreamLine
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string                        `json:"id"`
	Task       task.Task                     `json:"task"`
	Status     string                        `json:"status"` // waiting, running, success, failed
	Parameters map[string]velocity.Parameter `json:"parameters"`
}

func NewBuild(t task.Task, params map[string]velocity.Parameter) Build {
	return Build{
		ID:         uuid.NewV3(uuid.NewV1(), t.ID).String(),
		Task:       t,
		Status:     "waiting",
		Parameters: params,
	}
}

type BuildStep struct {
	velocity.Step
	ID            string                  `json:"id"`
	Number        uint64                  `json:"number"`
	Build         Build                   `json:"build"`
	Status        string                  `json:"status"` // waiting, running, success, failed
	OutputStreams []velocity.OutputStream `json:"outputStreams"`
}

func NewBuildStep(b Build, n uint64, step velocity.Step) BuildStep {
	bS := BuildStep{
		ID:     uuid.NewV3(uuid.NewV1(), b.ID).String(),
		Build:  b,
		Status: "waiting",
		Number: n,
		Step:   step,
	}

	outputStreams := []velocity.OutputStream{}
	for _, oS := range step.GetOutputStreams() {
		outputStreams = append(outputStreams, velocity.NewOutputStream(
			uuid.NewV3(uuid.NewV1(), bS.ID).String(),
			oS.Name,
		))
	}

	bS.OutputStreams = outputStreams

	return bS
}

type StreamLine struct {
	OutputStream velocity.OutputStream
	LineNumber   uint64
	Timestamp    time.Time
	Output       string
}

func NewStreamLine(oS velocity.OutputStream, lineNumber uint64, timestamp time.Time, output string) StreamLine {
	return StreamLine{
		OutputStream: oS,
		LineNumber:   lineNumber,
		Timestamp:    timestamp,
		Output:       output,
	}
}

type ResponseStreamLine struct {
	LineNumber uint64
	Timestamp  time.Time
	Output     string
}

type RequestBuild struct {
	TaskID     string             `json:"taskId"`
	Parameters []RequestParameter `json:"params"`
}

type RequestParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildManyResponse struct {
	Total  uint64          `json:"total"`
	Result []ResponseBuild `json:"result"`
}

type BuildStepManyResponse struct {
	Total  uint64              `json:"total"`
	Result []ResponseBuildStep `json:"result"`
}

type OutputStreamManyResponse struct {
	Total  uint64   `json:"total"`
	Result []string `json:"result"`
}

type ResponseBuild struct {
	ID      string                  `json:"id"`
	Project project.ResponseProject `json:"project"`
	Commit  commit.ResponseCommit   `json:"commit"`
	Task    task.ResponseTask       `json:"task"`
	Status  string                  `json:"status"`
}

func NewResponseBuild(b Build) ResponseBuild {
	return ResponseBuild{
		ID:      b.ID,
		Project: project.NewResponseProject(b.Task.Commit.Project),
		Commit:  commit.NewResponseCommit(b.Task.Commit),
		Task:    task.NewResponseTask(b.Task),
		Status:  b.Status,
	}
}

type ResponseBuildStep struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func NewResponseBuildStep(bS BuildStep) ResponseBuildStep {
	return ResponseBuildStep{
		ID:     bS.ID,
		Status: bS.Status,
	}
}
