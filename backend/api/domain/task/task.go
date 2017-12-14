package task

import (
	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	Create(t Task) Task
	Update(t Task) Task
	Delete(t Task)
	GetByTaskID(taskID string) (Task, error)
	GetByCommitIDAndTaskName(commitID string, name string) (Task, error)
	GetAllByCommitID(commitID string, q Query) ([]Task, uint64)
}

type Task struct {
	ID       string `json:"id"`
	CommitID string `json:"commitId"`
	velocity.Task
}

type Query struct {
	Amount uint64
	Page   uint64
}

type ResponseTask struct {
	ID string `json:"id"`
	velocity.Task
}

type ManyResponse struct {
	Total  uint64         `json:"total"`
	Result []ResponseTask `json:"result"`
}

func NewTask(commitID string, vTask velocity.Task) Task {
	return Task{
		ID:       uuid.NewV3(uuid.NewV1(), commitID).String(),
		CommitID: commitID,
		Task:     vTask,
	}
}

func NewResponseTask(t Task) ResponseTask {
	return ResponseTask{
		ID:   t.ID,
		Task: t.Task,
	}
}
