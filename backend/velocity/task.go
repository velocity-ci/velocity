package velocity

import (
	"encoding/json"
	"fmt"
)

type Task struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  map[string]Parameter `json:"parameters"`
	Steps       []Step               `json:"steps"`
}

func (t *Task) UpdateParams() {
	for _, s := range t.Steps {
		s.SetParams(t.Parameters)
	}
}

func (t *Task) String() string {
	j, _ := json.Marshal(t)
	return string(j)
}

func NewTask() Task {
	return Task{
		Name:        "",
		Description: "",
		Parameters:  map[string]Parameter{},
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
	var rawParameters map[string]*json.RawMessage
	err = json.Unmarshal(*objMap["parameters"], &rawParameters)
	if err != nil {
		return err
	}
	t.Parameters = make(map[string]Parameter)
	for paramName, rawMessage := range rawParameters {
		var p Parameter
		err = json.Unmarshal(*rawMessage, &p)
		p.Name = paramName
		t.Parameters[paramName] = p
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

		var s Step
		switch m["type"] {
		case "run":
			s = NewDockerRun()
			break
		case "build":
			s = NewDockerBuild()
			break
		case "clone":
			s = NewClone()
			break
		default:
			return fmt.Errorf("unsupported type in json.Unmarshal: %s", m["type"])
		}

		err := json.Unmarshal(*rawMessage, s)
		if err != nil {
			return err
		}
		t.Steps[index] = s
	}

	return nil
}
