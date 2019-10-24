package build

import (
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Step interface {
	GetID() string
	Execute(emitter Emitter, t *Task) error
	Stop() error
	GetType() string
	GetDescription() string
	GetDetails() string
	SetParams(map[string]*Parameter) error
	GetOutputStreams() []*Stream

	Validate(map[string]Parameter) error
}

// Step state constants
const (
	StateWaiting  = "waiting"
	StateBuilding = "building"
	StateSuccess  = "succeeded"
	StateFailed   = "failed"
)

//
// Event constants
// for Task_* we can add a modifier to specify *which* task e.g. TASK_COMPLETE-<task_name>
// for Step_* we can add a modifier to specify *which* step (in the currently building task) e.g. STEP_COMPLETE-<step_name>
const (
	EventBuildStart    = "BUILD_START"
	EventTaskStart     = "TASK_START"
	EventStepStart     = "STEP_START"
	EventStepComplete  = "STEP_COMPLETE" // fires regardless of success/fail
	EventStepSuccess   = "STEP_SUCCESS"
	EventStepFail      = "STEP_FAIL"
	EventTaskComplete  = "TASK_COMPLETE" // fires regardless of success/fail
	EventTaskSuccess   = "TASK_SUCCESS"
	EventTaskFail      = "TASK_FAIL"
	EventBuildComplete = "BUILD_COMPLETE" // fires regardless of success/fail
	EventBuildSuccess  = "BUILD_SUCCESS"
	EventBuildFail     = "BUILD_FAIL"
)

type Stream struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type BaseStep struct {
	ID          string `json:"id"`
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`

	OutputStreams []*Stream  `json:"outputStreams" yaml:"-"`
	Status        string     `json:"status"`
	StartedAt     *time.Time `json:"startedAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
	CompletedAt   *time.Time `json:"completedAt"`
}

func newBaseStep(t string, streamNames []string) BaseStep {
	streams := []*Stream{}
	for _, streamName := range streamNames {
		streams = append(streams, &Stream{
			ID:     uuid.NewV4().String(),
			Name:   streamName,
			Status: StateWaiting,
		})
	}
	return BaseStep{
		ID:            uuid.NewV4().String(),
		Type:          t,
		OutputStreams: streams,
		Status:        StateWaiting,
	}
}

func (bS *BaseStep) GetID() string {
	return bS.ID
}

func (bS *BaseStep) GetType() string {
	return bS.Type
}

func (bS *BaseStep) GetDescription() string {
	return bS.Description
}

func (bS *BaseStep) GetOutputStreams() []*Stream {
	return bS.OutputStreams
}

func (bS *BaseStep) GetStreamWriter(emitter Emitter, streamName string) (StreamWriter, error) {
	for _, outputStream := range bS.GetOutputStreams() {
		if outputStream.Name == streamName {
			return emitter.GetStreamWriter(outputStream), nil
		}
	}
	return nil, fmt.Errorf("could not find stream id for '%s'", streamName)
}

func (bS *BaseStep) GetStreamID(streamName string) (string, error) {
	for _, outputStream := range bS.GetOutputStreams() {
		if outputStream.Name == streamName {
			return outputStream.ID, nil
		}
	}
	return "", fmt.Errorf("could not find id for stream: %s", streamName)
}

type StreamLine struct {
	LineNumber uint64    `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}

func unmarshalStep(rawMessage []byte) (Step, error) {
	var m map[string]interface{}
	err := json.Unmarshal(rawMessage, &m)
	if err != nil {
		return nil, err
	}
	var s Step
	switch m["type"] {
	case "setup":
		s = &Setup{
			BaseStep: BaseStep{
				Type: "setup",
			},
		}
	case "run":
		s = &StepDockerRun{
			BaseStep: BaseStep{
				Type: "run",
			},
			Command: []string{},
		}
	case "build":
		s = &StepDockerBuild{
			BaseStep: BaseStep{
				Type: "build",
			},
		}
	case "compose":
		s = &StepDockerCompose{
			BaseStep: BaseStep{
				Type: "compose",
			},
		}
	case "push":
		s = &StepDockerPush{
			BaseStep: BaseStep{
				Type: "push",
			},
		}
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
