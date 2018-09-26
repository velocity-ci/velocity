package velocity

import (
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Step interface {
	Execute(emitter Emitter, t *Task) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
	GetOutputStreams() []string
	UnmarshalYamlInterface(map[interface{}]interface{}) error
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
	Type          string               `json:"type" yaml:"type"`
	Description   string               `json:"description" yaml:"description"`
	OutputStreams []string             `json:"outputStreams" yaml:"-"`
	Params        map[string]Parameter `json:"params" yaml:"-"`
	runID         string
}

func newBaseStep(t string, streams []string) BaseStep {
	return BaseStep{
		Type:          t,
		OutputStreams: streams,
		Params:        map[string]Parameter{},
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

func (bS *BaseStep) SetParams(params map[string]Parameter) {
	bS.Params = params
}

func (bS *BaseStep) GetRunID() string {
	if bS.runID == "" {
		// bS.runID = time.Now().Format("060102150405")
		bS.runID = uuid.NewV4().String()
	}

	return bS.runID
}

func (s *BaseStep) UnmarshalYamlInterface(y map[interface{}]interface{}) error {
	switch x := y["description"].(type) {
	case interface{}:
		s.Description = x.(string)
		break
	}
	return nil
}

type StreamLine struct {
	LineNumber uint64    `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}

func DetermineStepFromInterface(i map[string]interface{}) (Step, error) {
	switch i["type"] {
	case "setup":
		return NewSetup(), nil
	case "run":
		return NewDockerRun(), nil
	case "build":
		return NewDockerBuild(), nil
	case "compose":
		return NewDockerCompose(), nil
	case "push":
		return NewDockerPush(), nil
		// case "plugin":
		// 	var s Plugin
		// 	s.UnmarshalYamlInterface(y)
		// 	break
	}
	return nil, fmt.Errorf("could not determine step %+v", i)
}
