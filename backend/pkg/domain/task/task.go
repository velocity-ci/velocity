package task

import (
	"github.com/velocity-ci/velocity/backend/pkg/domain/githistory"
	"github.com/velocity-ci/velocity/backend/velocity"
)

type Task struct {
	UUID   string             `json:"id"`
	Commit *githistory.Commit `json:"commit"`
	*velocity.Task
}
