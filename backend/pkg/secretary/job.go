package secretary

import (
	"fmt"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
)

// EventPrefix - Global prefix for events.
const EventPrefix = "vlcty_"

// Architect -> Builder events
var (
	EventGetCommits = fmt.Sprintf("%sget-commits", EventPrefix)
)

type Job interface {
	GetID() string
	GetName() string
	Parse([]byte) error
	Do(*phoenix.Client) error
	Stop(*phoenix.Client) error
}

type baseJob struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (j *baseJob) GetID() string {
	return j.ID
}

func (j *baseJob) GetName() string {
	return j.Name
}

var jobs = []Job{
	NewSynchronise(),
}
