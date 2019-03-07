package task

import (
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/out"
)

type Step interface {
	Execute(emitter out.Emitter, t *Task) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
	GetOutputStreams() []string
	SetProjectRoot(string)
}

// Step state constants
const (
	StateWaiting = "waiting"
	StateRunning = "running"
	StateSuccess = "success"
	StateFailed  = "failed"
)

//
// Event constants
// for Task_* we can add a modifier to specify *which* task e.g. TASK_COMPLETE-<task_name>
// for Step_* we can add a modifier to specify *which* step (in the currently running task) e.g. STEP_COMPLETE-<step_name>
const (
	EventBuildStart    = "BUILD_START"
	EventTaskStart     = "TASK_START"
	EventStepStart     = "STEP_START"
	EventStepComplete  = "STEP_COMPLETE" // fires regardless of sucess/fail
	EventStepSuccess   = "STEP_SUCCESS"
	EventStepFail      = "STEP_FAIL"
	EventTaskComplete  = "TASK_COMPLETE" // fires regardless of success/fail
	EventTaskSuccess   = "TASK_SUCCESS"
	EventTaskFail      = "TASK_FAIL"
	EventBuildComplete = "BUILD_COMPLETE" // fires regardless of success/fail
	EventBuildSuccess  = "BUILD_SUCCESS"
	EventBuildFail     = "BUILD_FAIL"
)

type BaseStep struct {
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`

	OutputStreams []string  `json:"outputStreams" yaml:"-"`
	Status        string    `json:"status"`
	StartedAt     time.Time `json:"startedAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	CompletedAt   time.Time `json:"completedAt"`

	// Params      map[string]Parameter `json:"params" yaml:"-"`
	runID       string
	ProjectRoot string `json:"-"`
}

func newBaseStep(t string, streams []string) BaseStep {
	return BaseStep{
		Type:          t,
		OutputStreams: streams,
		Status:        StateWaiting,
		// Params:        map[string]Parameter{},
	}
}

func (bS *BaseStep) GetType() string {
	return bS.Type
}

func (bS *BaseStep) GetDescription() string {
	return bS.Description
}

func (bS *BaseStep) GetOutputStreams() []string {
	return bS.OutputStreams
}

// func (bS *BaseStep) SetParams(params map[string]Parameter) {
// 	bS.Params = params
// }

func (bS *BaseStep) GetRunID() string {
	if bS.runID == "" {
		bS.runID = uuid.NewV4().String()
	}

	return bS.runID
}

func (bS *BaseStep) SetProjectRoot(path string) {
	bS.ProjectRoot = path
}

type StreamLine struct {
	LineNumber uint64    `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}

func getStepFromBytes(rawMessage []byte) (Step, error) {
	var m map[string]interface{}
	err := json.Unmarshal(rawMessage, &m)
	if err != nil {
		return nil, err
	}
	var s Step
	switch m["type"] {
	case "setup":
		s = NewSetup()
	case "run":
		s = NewDockerRun()
	case "build":
		s = NewDockerBuild()
	case "compose":
		s = NewDockerCompose()
	case "push":
		s = NewDockerPush()
		// case "plugin":
		// 	s = NewPlugin()
		// 	break
	}

	if s == nil {
		return nil, fmt.Errorf("could not determine step %+v", m)
	}

	err = json.Unmarshal(rawMessage, s)
	return s, err
}
