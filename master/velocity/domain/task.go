package domain

import "time"

type Task struct {
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Project     Project     `json:"project" gorm:"primary_key"`
	Name        string      `json:"name" gorm:"primary_key"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	Steps       []Step
}

func (t *Task) UpdateParams() {
	for _, s := range t.Steps {
		s.SetParams(t.Parameters)
	}
}

type Parameter struct {
	Name         string   `json:"name" yaml:"name"`
	Value        string   `json:"default" yaml:"default"`
	OtherOptions []string `json:"otherOptions" yaml:"other_options"`
	Secret       bool     `json:"secret" yaml:"secret"`
}

type Step interface {
	Execute() error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate([]Parameter) error
	SetParams([]Parameter) error
}
