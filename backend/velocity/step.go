package velocity

import "time"

type Step interface {
	Execute(emitter Emitter, parameters map[string]Parameter) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
	GetOutputStreams() []string
}

type BaseStep struct {
	Type          string   `json:"type" yaml:"type"`
	Description   string   `json:"description" yaml:"description"`
	OutputStreams []string `json:"outputStreams" yaml:"-"`
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

type StreamLine struct {
	LineNumber uint64    `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Output     string    `json:"output"`
}
