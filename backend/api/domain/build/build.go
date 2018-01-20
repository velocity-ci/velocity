package build

import (
	"fmt"
	"time"

	"github.com/docker/go/canonical/json"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/api/domain/task"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	CreateBuild(b Build) Build
	UpdateBuild(b Build) Build
	DeleteBuild(b Build)
	GetBuildByBuildID(id string) (Build, error)
	GetBuildsByProjectID(projectID string, q BuildQuery) ([]Build, uint64)
	GetBuildsByCommitID(commitID string, q BuildQuery) ([]Build, uint64)
	GetBuildsByTaskID(taskID string, q BuildQuery) ([]Build, uint64)
	GetRunningBuilds() ([]Build, uint64)
	GetWaitingBuilds() ([]Build, uint64)

	// BuildSteps
	CreateBuildStep(bS BuildStep) BuildStep
	UpdateBuildStep(bS BuildStep) BuildStep
	DeleteBuildStep(bS BuildStep)
	GetBuildStepByBuildStepID(id string) (BuildStep, error)
	GetBuildStepByBuildIDAndNumber(buildID string, stepNumber uint64) (BuildStep, error)
	GetBuildStepsByBuildID(buildID string) ([]BuildStep, uint64)

	// OutputStreams
	SaveStream(s BuildStepStream) BuildStepStream
	GetStreamsByBuildStepID(buildStepID string) ([]BuildStepStream, uint64)
	GetStreamByID(id string) (BuildStepStream, error)

	GetStreamByBuildStepIDAndStreamName(buildStepID string, name string) (BuildStepStream, error)

	// StreamLines
	CreateStreamLine(sL StreamLine) StreamLine
	GetStreamLinesByStreamID(streamID string, q StreamLineQuery) ([]StreamLine, uint64)
}

type BuildQuery struct {
	Amount uint64
	Page   uint64
	Status string
}

type StreamLineQuery struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"projectId"`
	Task       task.Task         `json:"task"`
	Parameters map[string]string `json:"parameters"`

	Steps []BuildStep `json:"buildSteps"`

	Status      string    `json:"status"` // waiting, running, success, failed
	UpdatedAt   time.Time `json:"updatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func (b Build) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

func NewBuild(projectId string, t task.Task, params map[string]string) Build {
	return Build{
		ID:         uuid.NewV3(uuid.NewV1(), t.ID).String(),
		ProjectID:  projectId,
		Task:       t,
		Parameters: params,
		Status:     "waiting",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Steps:      []BuildStep{},
	}
}

type BuildStep struct {
	ID      string `json:"id"`
	BuildID string `json:"build"`
	Number  uint64 `json:"number"`

	Streams []BuildStepStream `json:"streams"`

	Status      string    `json:"status"` // waiting, running, success, failed
	UpdatedAt   time.Time `json:"updatedAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

type BuildStepStream struct {
	ID          string `json:"id"`
	BuildStepID string `json:"buildStepId"`
	Name        string `json:"name"`
}

func NewBuildStepStream(buildStepID string, name string) BuildStepStream {
	return BuildStepStream{
		ID:          uuid.NewV3(uuid.NewV1(), buildStepID).String(),
		BuildStepID: buildStepID,
		Name:        name,
	}
}

func NewBuildStep(buildID string, n uint64) BuildStep {
	bS := BuildStep{
		ID:        uuid.NewV3(uuid.NewV1(), buildID).String(),
		BuildID:   buildID,
		Status:    "waiting",
		Number:    n,
		UpdatedAt: time.Now(),
		Streams:   []BuildStepStream{},
	}

	return bS
}

type StreamLine struct {
	BuildStepStreamID string    `json:"streamId"`
	LineNumber        uint64    `json:"lineNumber"`
	Timestamp         time.Time `json:"timestamp"`
	Output            string    `json:"output"`
}

func NewStreamLine(buildStepStreamID string, lineNumber uint64, timestamp time.Time, output string) StreamLine {
	return StreamLine{
		BuildStepStreamID: buildStepStreamID,
		LineNumber:        lineNumber,
		Timestamp:         timestamp,
		Output:            fmt.Sprintf("%s\n", output),
	}
}

type ResponseStreamLine struct {
	LineNumber uint64
	Timestamp  time.Time
	Output     string
}

type RequestBuild struct {
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

type ResponseOutputStream struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OutputStreamManyResponse struct {
	Total  uint64                 `json:"total"`
	Result []ResponseOutputStream `json:"result"`
}

type StreamLineManyResponse struct {
	Total  uint64       `json:"total"`
	Result []StreamLine `json:"result"`
}

type ResponseBuild struct {
	ID          string              `json:"id"`
	TaskID      string              `json:"task"`
	Status      string              `json:"status"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	CreatedAt   time.Time           `json:"createdAt"`
	StartedAt   time.Time           `json:"startedAt"`
	CompletedAt time.Time           `json:"completedAt"`
	Steps       []ResponseBuildStep `json:"steps"`
}

func NewResponseBuild(b Build) ResponseBuild {
	steps := []ResponseBuildStep{}
	for i, s := range b.Steps {
		steps = append(steps, NewResponseBuildStep(s, b.Task.Steps[i]))
	}
	return ResponseBuild{
		ID:          b.ID,
		TaskID:      b.Task.ID,
		Status:      b.Status,
		UpdatedAt:   b.UpdatedAt,
		CreatedAt:   b.CreatedAt,
		StartedAt:   b.StartedAt,
		CompletedAt: b.CompletedAt,
		Steps:       steps,
	}
}

type ResponseBuildStep struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Number      uint64                 `json:"number"`
	Status      string                 `json:"status"`
	StartedAt   time.Time              `json:"startedAt"`
	CompletedAt time.Time              `json:"completedAt"`
	Streams     []ResponseOutputStream `json:"streams"`
}

func NewResponseBuildStep(bS BuildStep, s velocity.Step) ResponseBuildStep {
	streams := []ResponseOutputStream{}
	for _, s := range bS.Streams {
		streams = append(streams, ResponseOutputStream{
			ID:   s.ID,
			Name: s.Name,
		})
	}
	return ResponseBuildStep{
		ID:          bS.ID,
		Type:        s.GetType(),
		Description: s.GetDescription(),
		Number:      bS.Number,
		Status:      bS.Status,
		StartedAt:   bS.StartedAt,
		CompletedAt: bS.CompletedAt,
		Streams:     streams,
	}
}

type WebsocketBuildStep struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func NewWebsocketBuildStep(bS BuildStep) WebsocketBuildStep {
	return WebsocketBuildStep{
		ID:          bS.ID,
		Status:      bS.Status,
		StartedAt:   bS.StartedAt,
		CompletedAt: bS.CompletedAt,
	}
}
