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
	SaveBuild(b *Build) *Build
	DeleteBuild(b *Build)
	GetBuildByProjectAndCommitAndID(p *project.Project, c *commit.Commit, id string) (*Build, error)
	// Order timestamp descending
	GetBuildsByProject(p *project.Project, q Query) ([]*Build, uint64)
	// Order timestamp descending
	GetBuildsByProjectAndCommit(p *project.Project, c *commit.Commit) ([]*Build, uint64)
	GetRunningBuilds() ([]*Build, uint64)
	GetWaitingBuilds() ([]*Build, uint64)

	// BuildSteps
	SaveBuildStep(bS *BuildStep) *BuildStep
	GetBuildStepsForBuild(b *Build) ([]*BuildStep, uint64)
	GetBuildStepByBuildAndID(b *Build, id string) (*BuildStep, error)

	// OutputStreams
	SaveOutputStream(oS *OutputStream) *OutputStream
	GetOutputStreamsForBuildStep(bS *BuildStep) ([]*OutputStream, uint64)
	GetOutputStreamByID(id string) (*OutputStream, error)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string
	Task       task.Task
	Status     string // waiting, running, success, failed
	Parameters map[string]velocity.Parameter
}

func NewBuild(t *task.Task, params map[string]velocity.Parameter) *Build {
	return &Build{
		ID:         uuid.NewV3(uuid.NewV1(), t.ID).String(),
		Task:       *t,
		Status:     "waiting",
		Parameters: params,
	}
}

type SlaveBuild struct {
	ID         string                        `json:"id"`
	Task       task.Task                     `json:"task"`
	Status     string                        `json:"status"`
	Parameters map[string]velocity.Parameter `json:"parameters"`
}

type BuildStep struct {
	ID          string
	Number      uint64
	Description string
	Build       Build
	Status      string // waiting, running, success, failed
	Step        velocity.Step
}

func NewBuildStep(b *Build, n uint64, desc string, step velocity.Step) *BuildStep {
	return &BuildStep{
		ID:          uuid.NewV3(uuid.NewV1(), b.ID).String(),
		Build:       *b,
		Status:      "waiting",
		Number:      n,
		Description: desc,
		Step:        step,
	}
}

type OutputStream struct {
	ID        string
	BuildStep BuildStep
	Name      string
	Path      string
}

func NewOutputStream(bS *BuildStep, name string) *OutputStream {
	return &OutputStream{
		ID:        uuid.NewV3(uuid.NewV1(), bS.ID).String(),
		BuildStep: *bS,
		Name:      name,
	}
}

type ResponseOutputStream struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewResponseOutputStream(oS *OutputStream) *ResponseOutputStream {
	return &ResponseOutputStream{
		ID:   oS.ID,
		Name: oS.Name,
	}
}

type StreamLine struct {
	OutputStream OutputStream
	LineNumber   uint64
	Timestamp    time.Time
	Output       string
}

func NewStreamLine(oS *OutputStream, lineNumber uint64, timestamp time.Time, output string) *StreamLine {
	return &StreamLine{
		OutputStream: *oS,
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
	Total  uint64           `json:"total"`
	Result []*ResponseBuild `json:"result"`
}

type BuildStepManyResponse struct {
	Total  uint64               `json:"total"`
	Result []*ResponseBuildStep `json:"result"`
}

type OutputStreamManyResponse struct {
	Total  uint64
	Result []*ResponseOutputStream `json:"result"`
}

type ResponseBuild struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project"`
	CommitHash string `json:"commit"`
	TaskID     string `json:"taskID"`
	Status     string `json:"status"`
}

func NewResponseBuild(b *Build) *ResponseBuild {
	return &ResponseBuild{
		ID:         b.ID,
		ProjectID:  b.Task.Commit.Project.ID,
		CommitHash: b.Task.Commit.ID,
		TaskID:     b.Task.ID,
		Status:     b.Status,
	}
}

type ResponseBuildStep struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func NewResponseBuildStep(bS *BuildStep) *ResponseBuildStep {
	return &ResponseBuildStep{
		ID:     bS.ID,
		Status: bS.Status,
	}
}
