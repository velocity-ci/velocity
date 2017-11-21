package task

import (
	"fmt"

	"github.com/velocity-ci/velocity/backend/api/domain/commit"
	"github.com/velocity-ci/velocity/backend/api/domain/project"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Repository interface {
	Save(t *Task) *Task
	Delete(t *Task)
	GetByProjectAndCommitAndID(p *project.Project, c *commit.Commit, ID string) (*Task, error)
	GetAllByProjectAndCommit(p *project.Project, c *commit.Commit, q Query) ([]*Task, uint64)
}

type Task struct {
	ID     string
	Commit commit.Commit
	VTask  velocity.Task
}

type Query struct {
	Amount uint64
	Page   uint64
}

type ManyResponse struct {
	Total  uint64  `json:"total"`
	Result []*Task `json:"result"`
}

func NewTask(p *project.Project, c *commit.Commit, vTask velocity.Task) *Task {
	return &Task{
		ID:     fmt.Sprintf("%s-%s-%s", p.ID, c.Hash[:7], vTask.Name),
		Commit: *c,
		VTask:  vTask,
	}
}
