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

func (t *Task) SetEmitter(e func(string)) {
	for _, s := range t.Steps {
		s.SetEmitter(e)
	}
}
