package commit

import "github.com/velocity-ci/velocity/backend/task"

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
	ID         uint64     `json:"id"`
	ProjectID  string     `json:"project"`
	CommitHash string     `json:"commit"`
	Task       *task.Task `json:"task"`
	Status     string     `json:"status"`
}

type QueuedBuild struct {
	ProjectID  string `json:"project"`
	CommitHash string `json:"commit"`
	ID         uint64 `json:"id"`
	Status     string `json:"status"`
}

func NewBuild(projectID string, commitHash string, t *task.Task) Build {
	return Build{
		ProjectID:  projectID,
		CommitHash: commitHash,
		Task:       t,
		Status:     "waiting",
	}
}

func NewResponseBuild(b *Build) *ResponseBuild {
	return &ResponseBuild{
		ID:         b.ID,
		ProjectID:  b.ProjectID,
		CommitHash: b.CommitHash,
		TaskName:   b.Task.Name,
		Status:     b.Status,
	}
}

func NewQueuedBuild(b *Build) *QueuedBuild {
	return &QueuedBuild{
		ProjectID:  b.ProjectID,
		CommitHash: b.CommitHash,
		ID:         b.ID,
		Status:     b.Status,
	}
}
