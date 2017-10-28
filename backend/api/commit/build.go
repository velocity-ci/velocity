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

type Build struct {
	ProjectID  string     `json:"project"`
	CommitHash string     `json:"commit"`
	Task       *task.Task `json:"task"`
	Status     string     `json:"status"`
}

func NewBuild(projectID string, commitHash string, t *task.Task) Build {
	return Build{
		ProjectID:  projectID,
		CommitHash: commitHash,
		Task:       t,
		Status:     "waiting",
	}
}
