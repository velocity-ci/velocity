package velocity

type Step interface {
	Execute(emitter Emitter, parameters map[string]Parameter) error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate(map[string]Parameter) error
	SetParams(map[string]Parameter) error
}

type BaseStep struct {
	Type        string `json:"type" yaml:"type"`
	Description string `json:"description" yaml:"description"`
}
