package build

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	CreateBuild(b Build) Build
	UpdateBuild(b Build) Build
	DeleteBuild(b Build)
	GetBuildByBuildID(id string) (Build, error)
	GetBuildsByProjectID(projectID string, q Query) ([]Build, uint64)
	GetBuildsByCommitID(commitID string, q Query) ([]Build, uint64)
	GetBuildsByTaskID(taskID string, q Query) ([]Build, uint64)
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
	GetStreamLinesByStreamID(streamID string) ([]StreamLine, uint64)
}

type Query struct {
	Amount uint64
	Page   uint64
}

type Build struct {
	ID         string                        `json:"id"`
	ProjectID  string                        `json:"projectId"`
	TaskID     string                        `json:"taskId"`
	Parameters map[string]velocity.Parameter `json:"parameters"`

	Status      string    `json:"status"` // waiting, running, success, failed
	CreatedAt   time.Time `json:"createdAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func NewBuild(projectId string, taskID string, params map[string]velocity.Parameter) Build {
	return Build{
		ID:         uuid.NewV3(uuid.NewV1(), taskID).String(),
		ProjectID:  projectId,
		TaskID:     taskID,
		Parameters: params,
		Status:     "waiting",
		CreatedAt:  time.Now(),
	}
}

type BuildStep struct {
	ID      string `json:"id"`
	BuildID string `json:"build"`
	Number  uint64 `json:"number"`

	Status      string    `json:"status"` // waiting, running, success, failed
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
		ID:      uuid.NewV3(uuid.NewV1(), buildID).String(),
		BuildID: buildID,
		Status:  "waiting",
		Number:  n,
	}

	return bS
}

type StreamLine struct {
	BuildStepStreamID string
	LineNumber        uint64
	Timestamp         time.Time
	Output            string
}

func NewStreamLine(buildStepStreamID string, lineNumber uint64, timestamp time.Time, output string) StreamLine {
	return StreamLine{
		BuildStepStreamID: buildStepStreamID,
		LineNumber:        lineNumber,
		Timestamp:         timestamp,
		Output:            output,
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

type OutputStreamManyResponse struct {
	Total  uint64   `json:"total"`
	Result []string `json:"result"`
}

type StreamLineManyResponse struct {
	Total  uint64       `json:"total"`
	Result []StreamLine `json:"result"`
}

type ResponseBuild struct {
	ID     string `json:"id"`
	TaskID string `json:"task"`
	Status string `json:"status"`
}

func NewResponseBuild(b Build) ResponseBuild {
	return ResponseBuild{
		ID:     b.ID,
		TaskID: b.TaskID,
		Status: b.Status,
	}
}

type ResponseBuildStep struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Number      uint64    `json:"number"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func NewResponseBuildStep(bS BuildStep, s velocity.Step) ResponseBuildStep {
	return ResponseBuildStep{
		ID:          bS.ID,
		Type:        s.GetType(),
		Description: s.GetDescription(),
		Number:      bS.Number,
		Status:      bS.Status,
		StartedAt:   bS.StartedAt,
		CompletedAt: bS.CompletedAt,
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
