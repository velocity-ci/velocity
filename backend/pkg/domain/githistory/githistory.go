package githistory

import (
	"time"

	"github.com/velocity-ci/velocity/backend/pkg/domain/project"
)

type Commit struct {
	UUID      string           `json:"id"`
	Project   *project.Project `json:"project"`
	Hash      string           `json:"hash"`
	Author    string           `json:"author"`
	CreatedAt time.Time        `json:"createdAt"`
	Message   string           `json:"message"`
	Branches  []*Branch        `json:"branches"`
}

type Branch struct {
	UUID        string           `json:"id"`
	Name        string           `json:"name"`
	Project     *project.Project `json:"projectId"`
	LastUpdated time.Time        `json:"lastUpdated"`
	Active      bool             `json:"active"`
}
