package task

import (
	"encoding/json"
	"fmt"
)

type Task struct {
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

func NewTask() Task {
	return Task{
		Name:        "",
		Description: "",
		Parameters:  []Parameter{},
		Steps:       []Step{},
	}
}

func (t *Task) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Name
	err = json.Unmarshal(*objMap["name"], &t.Name)
	if err != nil {
		return err
	}

	// Deserialize Description
	err = json.Unmarshal(*objMap["description"], &t.Description)
	if err != nil {
		return err
	}

	// Deserialize Parameters
	var rawParameters []*json.RawMessage
	err = json.Unmarshal(*objMap["parameters"], &rawParameters)
	if err != nil {
		return err
	}
	t.Parameters = make([]Parameter, len(rawParameters))
	for index, rawMessage := range rawParameters {
		var p Parameter
		err = json.Unmarshal(*rawMessage, &p)
		t.Parameters[index] = p
	}

	// Deserialize Steps by type
	var rawSteps []*json.RawMessage
	err = json.Unmarshal(*objMap["steps"], &rawSteps)
	if err != nil {
		return err
	}
	t.Steps = make([]Step, len(rawSteps))
	var m map[string]interface{}
	for index, rawMessage := range rawSteps {
		err = json.Unmarshal(*rawMessage, &m)
		if err != nil {
			return err
		}

		if m["type"] == "run" {
			s := NewDockerRun()
			err := json.Unmarshal(*rawMessage, &s)
			if err != nil {
				return err
			}
			t.Steps[index] = &s
		} else if m["type"] == "build" {
			s := NewDockerBuild()
			err := json.Unmarshal(*rawMessage, &s)
			if err != nil {
				return err
			}
			t.Steps[index] = &s
		} else {
			return fmt.Errorf("unsupported type in json.Unmarshal: %s", m["type"])
		}
	}

	return nil
}
