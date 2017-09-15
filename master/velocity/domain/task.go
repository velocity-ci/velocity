package domain

import "time"

type Task struct {
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Project     Project     `json:"project"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	Steps       []Step      `json:"steps"`
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
