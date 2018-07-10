package velocity

import (
	"encoding/json"

	"go.uber.org/zap"
)

type Task struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Git         TaskGit           `json:"git" yaml:"git"`
	Docker      TaskDocker        `json:"docker" yaml:"docker"`
	Parameters  []ParameterConfig `json:"parameters" yaml:"parameters"`
	Steps       []Step            `json:"steps" yaml:"steps"`

	RunID              string               `json:"-" yaml:"-"`
	ResolvedParameters map[string]Parameter `json:"-" yaml:"-"`
}

type TaskGit struct {
	Submodule bool `json:"submodule"`
}

type TaskDocker struct {
	Registries []DockerRegistry `json:"registries" yaml:"registries"`
}

type DockerRegistry struct {
	Address            string            `json:"address"`
	Use                string            `json:"use" yaml:"use"`
	Arguments          map[string]string `json:"arguments" yaml:"arguments"`
	AuthorizationToken string            `json:"authToken"`
}

func (d *DockerRegistry) UnmarshalYamlInterface(y map[interface{}]interface{}) error {

	switch x := y["use"].(type) {
	case interface{}:
		d.Use = x.(string)
		break
	}

	d.Arguments = map[string]string{}
	switch x := y["arguments"].(type) {
	case map[interface{}]interface{}:
		for k, v := range x {
			d.Arguments[k.(string)] = v.(string)
		}
		break
	}
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
		Parameters:  []ParameterConfig{},
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
	json.Unmarshal(*objMap["name"], &t.Name)

	// Deserialize Description
	json.Unmarshal(*objMap["description"], &t.Description)

	// Deserialize Parameters
	if val, _ := objMap["parameters"]; val != nil {
		var rawParameters []*json.RawMessage
		err = json.Unmarshal(*val, &rawParameters)
		if err == nil {
			t.Parameters = []ParameterConfig{}
			for _, rawMessage := range rawParameters {
				var m map[string]interface{}
				err = json.Unmarshal(*rawMessage, &m)
				if err != nil {
					GetLogger().Error("could not unmarshal parameters", zap.Error(err))
					return err
				}
				if _, ok := m["use"]; ok { // derivedParam
					p := DerivedParameter{}
					err = json.Unmarshal(*rawMessage, &p)
					if err != nil {
						GetLogger().Error("could not unmarshal determined parameter", zap.Error(err))
						return err
					}
					t.Parameters = append(t.Parameters, p)
				} else if _, ok := m["name"]; ok { // basicParam
					p := BasicParameter{}
					err = json.Unmarshal(*rawMessage, &p)
					if err != nil {
						GetLogger().Error("could not unmarshal determined parameter", zap.Error(err))
						return err
					}
					t.Parameters = append(t.Parameters, p)
				}

			}
		}
	}

	t.Docker = TaskDocker{}
	json.Unmarshal(*objMap["docker"], &t.Docker)

	// Deserialize Steps by type
	if val, _ := objMap["steps"]; val != nil {
		var rawSteps []*json.RawMessage
		err = json.Unmarshal(*val, &rawSteps)
		if err == nil {
			t.Steps = []Step{}
			var m map[string]interface{}
			for _, rawMessage := range rawSteps {
				err = json.Unmarshal(*rawMessage, &m)
				if err != nil {
					GetLogger().Error("could not unmarshal step", zap.Error(err))

					return err
				}

				s, err := DetermineStepFromInterface(m)
				if err != nil {
					GetLogger().Error("could not determine step from interface", zap.Error(err))
				} else {
					err := json.Unmarshal(*rawMessage, s)
					if err != nil {
						GetLogger().Error("could not unmarshal step", zap.Error(err))
					} else {
						t.Steps = append(t.Steps, s)
					}
				}
			}
		}
	}

	return nil
}

func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var taskMap map[string]interface{}
	err := unmarshal(&taskMap)
	if err != nil {
		GetLogger().Error("could not unmarshal task", zap.Error(err))
		return err
	}

	switch x := taskMap["name"].(type) {
	case string:
		t.Name = x
		break
	}

	switch x := taskMap["description"].(type) {
	case string:
		t.Description = x
		break
	}

	t.Git = TaskGit{
		Submodule: false,
	}
	switch x := taskMap["git"].(type) {
	case map[interface{}]interface{}:
		t.Git = TaskGit{
			Submodule: x["submodule"].(bool),
		}
		break
	}

	t.Docker = TaskDocker{
		Registries: []DockerRegistry{},
	}
	switch x := taskMap["docker"].(type) {
	case map[interface{}]interface{}:
		switch y := x["registries"].(type) {
		case []interface{}:
			for _, r := range y {
				switch z := r.(type) {
				case map[interface{}]interface{}:
					d := DockerRegistry{}
					d.UnmarshalYamlInterface(z)
					t.Docker.Registries = append(t.Docker.Registries, d)
				}
			}
			break
		}
		break
	}

	t.Parameters = unmarshalConfigParameters(taskMap["parameters"])

	t.Steps = []Step{}
	switch x := taskMap["steps"].(type) {
	case []interface{}:
		for _, s := range x {
			switch y := s.(type) {
			case map[interface{}]interface{}:
				m := map[string]interface{}{} // generate map[string]interface{}
				for k, v := range y {
					m[k.(string)] = v
				}
				s, err := DetermineStepFromInterface(m)
				if err != nil {
					GetLogger().Error("could not determine step from interface", zap.Error(err))
				} else {
					err = s.UnmarshalYamlInterface(y)
					if err != nil {
						GetLogger().Error("could not unmarshal yaml step", zap.Error(err))
					} else {
						t.Steps = append(t.Steps, s)
					}
				}
				break
			}
		}
		break
	}

	return nil
}
