package velocity

import (
	"fmt"
	"time"
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

const (
	successANSI = "\x1b[1m\x1b[49m\x1b[32m"
	errorANSI   = "\x1b[1m\x1b[49m\x1b[31m"
	infoANSI    = "\x1b[1m\x1b[49m\x1b[34m"
)

// Step state constants
const (
	StateWaiting = "waiting"
	StateRunning = "running"
	StateSuccess = "success"
	StateFailed  = "failed"
)

type BaseStep struct {
	Type          string               `json:"type" yaml:"type"`
	Description   string               `json:"description" yaml:"description"`
	OutputStreams []string             `json:"outputStreams" yaml:"-"`
	Params        map[string]Parameter `json:"params" yaml:"-"`
	runID         string
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
		bS.runID = time.Now().Format("060102150405")
	}

	return bS.runID
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
