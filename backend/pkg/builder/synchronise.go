package builder

import (
	"encoding/json"
	"fmt"

	"github.com/velocity-ci/velocity/backend/pkg/phoenix"
)

// Builder -> Architect (Synchronise events)
var (
	eventSynchronisePrefix    = fmt.Sprintf("%ssynchronise-", EventPrefix)
	EventSynchroniseNewCommit = fmt.Sprintf("%snew-commit", eventSynchronisePrefix)
	EventSynchroniseNewTask   = fmt.Sprintf("%snew-task", eventSynchronisePrefix)
	EventSynchroniseNewBranch = fmt.Sprintf("%snew-branch", eventSynchronisePrefix)
)

type Synchronise struct {
	*baseJob
	Project string `json:"project"`
}

func NewSynchronise() *Synchronise {
	return &Synchronise{
		baseJob: &baseJob{
			Name: "synchronise",
		},
	}
}

func (j *Synchronise) Parse(payloadBytes []byte) error {
	return json.Unmarshal(payloadBytes, j)
}

func (j *Synchronise) Do(ws *phoenix.Client) error {

	return nil
}

func (j *Synchronise) Stop(ws *phoenix.Client) error {

	return nil
}
