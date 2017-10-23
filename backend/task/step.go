package task

type Step interface {
	Execute(map[string]Parameter) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
	SetEmitter(func(string))
}

type BaseStep struct {
	Type        string       `json:"type" yaml:"type"`
	Description string       `json:"description" yaml:"description"`
	Emit        func(string) `json:"-" yaml:"-"`
}
