package task

type Step interface {
	Execute(step uint64, parameters map[string]Parameter) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
	SetEmitter(func(status string, step uint64, output string))
}

type BaseStep struct {
	Type        string                                          `json:"type" yaml:"type"`
	Description string                                          `json:"description" yaml:"description"`
	Emit        func(status string, step uint64, output string) `json:"-" yaml:"-"`
}
