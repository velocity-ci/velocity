package task

type Parameter struct {
	Name         string   `json:"name" yaml:"name"`
	Value        string   `json:"default" yaml:"default"`
	OtherOptions []string `json:"otherOptions" yaml:"other_options"`
	Secret       bool     `json:"secret" yaml:"secret"`
}