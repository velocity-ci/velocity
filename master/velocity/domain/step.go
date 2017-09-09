package domain

type Step interface {
	Execute() error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate([]Parameter) error
	SetParams([]Parameter) error
	SetEmitter(func(string))
}

type BaseStep struct {
	Type       string `json:"type" yaml:"type"`
	Parameters []Parameter
	Emit       func(string)
}
