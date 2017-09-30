package task

type Step interface {
	Execute([]Parameter) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate([]Parameter) error
	SetParams([]Parameter) error
	SetEmitter(func(string))
}

type BaseStep struct {
	Type        string       `json:"type" yaml:"type"`
	Description string       `json:"description" yaml:"description"`
	Emit        func(string) `json:"-" yaml:"-"`
}
