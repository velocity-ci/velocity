package velocity

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Task struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Git         TaskGit           `json:"git"`
	Parameters  []ConfigParameter `json:"parameters"`
	Steps       []Step            `json:"steps"`
	runID       string
}

type TaskGit struct {
	Submodule bool `json:"submodule"`
}

// Maybe pass git repository to clone?
func (t *Task) Setup(emitter Emitter) error {
	t.runID = fmt.Sprintf("vci-%s", time.Now().Format("060102150405"))

	writer := emitter.GetStreamWriter("setup")
	writer.SetStatus(StateRunning)

	// Resolve parameters
	parameters := map[string]Parameter{}
	for _, config := range t.Parameters {
		params, err := config.GetParameters(writer, t.runID)
		if err != nil {
			writer.SetStatus(StateFailed)
			log.Printf("could not resolve parameter: %v", err)
		}
		for _, param := range params {
			parameters[param.Name] = param
		}
	}

	// Update params on steps
	for _, s := range t.Steps {
		s.SetParams(parameters)
	}

	writer.SetStatus(StateSuccess)
	writer.Write([]byte("\nSetup success.\n"))

	return nil
}

func (t *Task) String() string {
	j, _ := json.Marshal(t)
	return string(j)
}

func NewTask() Task {
	return Task{
		Name:        "",
		Description: "",
		Parameters:  []ConfigParameter{},
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
	t.Parameters = []ConfigParameter{}
	for _, rawMessage := range rawParameters {
		var p ConfigParameter
		err = json.Unmarshal(*rawMessage, &p)
		// p.Name = paramName
		t.Parameters = append(t.Parameters, p)
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
