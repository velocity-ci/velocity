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

type Parameter struct {
	Name    string `json:"name" yaml:"name"`
	Default string `json:"default" yaml:"default"`
}

type Step interface {
	Execute() error
	GetType() string
	GetDescription() string
	GetDetails() string
	Validate([]Parameter) error
}
