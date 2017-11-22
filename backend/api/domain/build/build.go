package build

import (
	"fmt"
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

	// BuildSteps
	SaveBuildStep(bS *BuildStep) *BuildStep
	GetBuildStepsForBuild(b *Build) ([]*BuildStep, uint64)
	GetBuildStepByBuildAndID(b *Build, id string) (*BuildStep, error)

	// OutputStreams
	SaveOutputStream(oS *OutputStream) *OutputStream
	GetOutputStreamsForBuildStep(bS *BuildStep) ([]*OutputStream, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string
	Project    project.Project
	Commit     commit.Commit
	Task       task.Task
	Status     string
	Parameters map[string]velocity.Parameter
}

func NewBuild(p *project.Project, c *commit.Commit, t *task.Task, params map[string]velocity.Parameter) *Build {
	return &Build{
		ID:         uuid.NewV3(uuid.NewV1(), fmt.Sprintf("%s-%s-%s", p.ID, c.Hash[:7], t.ID)).String(),
		Project:    *p,
		Commit:     *c,
		Task:       *t,
		Status:     "",
		Parameters: params,
	}
}

type BuildStep struct {
	ID     string
	Build  Build
	Status string
}

func NewBuildStep(b *Build) *BuildStep {
	return &BuildStep{
		ID:     uuid.NewV3(uuid.NewV1(), b.ID).String(),
		Build:  *b,
		Status: "",
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
		Path:      "",
	}
}

type StreamLine struct {
	OutputStreamID string
	LineNumber     uint64
	Timestamp      time.Time
	Output         string
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
		ProjectID:  b.Project.ID,
		CommitHash: b.Commit.ID,
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
