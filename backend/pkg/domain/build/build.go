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

type BuildQuery struct {
	Limit  int    `json:"amount" query:"amount"`
	Page   int    `json:"page" query:"page"`
	Status string `json:"status" query:"status"`
}

func NewBuildQuery() *BuildQuery {
	return &BuildQuery{
		Limit: 10,
		Page:  1,
	}
}
