package commit

import (
	"time"

	"github.com/velocity-ci/velocity/backend/velocity"
)

type RequestBuild struct {
	TaskName   string             `json:"taskName"`
	Parameters []RequestParameter `json:"params"`
}

type RequestParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseBuild struct {
	ID         uint64 `json:"id"`
	ProjectID  string `json:"project"`
	CommitHash string `json:"commit"`
	TaskName   string `json:"taskName"`
	Status     string `json:"status"`
}

type Build struct {
	ID       uint64         `json:"id"`
	Task     *velocity.Task `json:"task"`
	Status   string         `json:"status"`
	StepLogs []StepLog      `json:"logs"`
}

type StepLog struct {
	Logs   map[string][]Log `json:"containerLogs"` //containerName: logs
	Status string           `json:"status"`
}

type Log struct {
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

type QueuedBuild struct {
	ProjectID  string `json:"project"`
	CommitHash string `json:"commit"`
	ID         uint64 `json:"id"`
}

func NewBuild(projectID string, commitHash string, t *velocity.Task) Build {
	return Build{
		Task:     t,
		Status:   "waiting",
		StepLogs: []StepLog{},
	}
}

func NewResponseBuild(b *Build, projectID string, commitHash string) *ResponseBuild {
	return &ResponseBuild{
		ID:         b.ID,
		ProjectID:  projectID,
		CommitHash: commitHash,
		TaskName:   b.Task.Name,
		Status:     b.Status,
	}
}

func NewQueuedBuild(b *Build, projectID string, commitHash string) *QueuedBuild {
	return &QueuedBuild{
		ProjectID:  projectID,
		CommitHash: commitHash,
		ID:         b.ID,
	}
}
