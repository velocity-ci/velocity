package config

import (
	"encoding/json"
)

type Task struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Docker      TaskDocker  `json:"docker"`
	Parameters  []Parameter `json:"parameters"`
	Steps       []Step      `json:"steps"`

	ParseErrors      []string `json:"parseErrors"`
	ValidationErrors []string `json:"validationErrors"`
}

func NewTask() *Task {
	return &Task{
		Name:        "",
		Description: "",
		Docker: TaskDocker{
			Registries: []TaskDockerRegistry{},
		},
		Parameters:       []Parameter{},
		Steps:            []Step{},
		ParseErrors:      []string{},
		ValidationErrors: []string{},
	}
}

func handleUnmarshalError(t *Task, err error) *Task {
	if err != nil {
		t.ParseErrors = append(t.ParseErrors, err.Error())
	}

	return t
}

func (t *Task) UnmarshalJSON(b []byte) error {
	// We don't return any errors from this function so we can show more helpful parse errors
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		t = handleUnmarshalError(t, err)
	}

	// Deserialize Name TODO: remove
	if _, ok := objMap["name"]; ok {
		err = json.Unmarshal(*objMap["name"], &t.Name)
		t = handleUnmarshalError(t, err)
	}

	// Deserialize Description
	if _, ok := objMap["description"]; ok {
		err = json.Unmarshal(*objMap["description"], &t.Description)
		t = handleUnmarshalError(t, err)
	}

	// Deserialize Parameters
	if val, _ := objMap["parameters"]; val != nil {
		var rawParameters []*json.RawMessage
		err = json.Unmarshal(*val, &rawParameters)
		t = handleUnmarshalError(t, err)
		if err == nil {
			for _, rawMessage := range rawParameters {
				param, err := unmarshalParameter(*rawMessage)
				t = handleUnmarshalError(t, err)
				if param != nil {
					t.Parameters = append(t.Parameters, param)
				}
			}
		}
	}

	// Deserialize Docker
	if _, ok := objMap["docker"]; ok {
		err = json.Unmarshal(*objMap["docker"], &t.Docker)
		t = handleUnmarshalError(t, err)
	}

	// Deserialize Steps by type
	if val, _ := objMap["steps"]; val != nil {
		var rawSteps []*json.RawMessage
		err = json.Unmarshal(*val, &rawSteps)
		t = handleUnmarshalError(t, err)
		if err == nil {
			for _, rawMessage := range rawSteps {
				s, err := unmarshalStep(*rawMessage)
				t = handleUnmarshalError(t, err)
				if err == nil {
					err = json.Unmarshal(*rawMessage, s)
					t = handleUnmarshalError(t, err)
					if err == nil {
						t.Steps = append(t.Steps, s)
					}
				}
			}
		}
	}

	return nil
}
