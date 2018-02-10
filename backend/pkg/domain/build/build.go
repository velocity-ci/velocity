package build

import (
	"encoding/json"
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/task"
)

type Build struct {
	ID         string            `json:"id"`
	Task       *task.Task        `json:"task"`
	Parameters map[string]string `json:"parameters"`

	// Steps []*Step `json:"buildSteps"`

	Status string `json:"status"`

	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	StartedAt   time.Time `json:"startedAt"`
	CompletedAt time.Time `json:"completedAt"`
}

func (s Build) String() string {
	j, _ := json.Marshal(s)
	return string(j)
}
