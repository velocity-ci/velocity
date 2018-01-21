package velocity

import (
	"encoding/json"
	"log"
)

type Task struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Git         TaskGit           `json:"git" yaml:"git"`
	Parameters  []ConfigParameter `json:"parameters" yaml:"parameters"`
	Steps       []Step            `json:"steps" yaml:"steps"`
	runID       string
}

type TaskGit struct {
	Submodule bool `json:"submodule"`
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
	var rawParameters []*json.RawMessage
	err = json.Unmarshal(*objMap["parameters"], &rawParameters)
	if err != nil {
		log.Println("could not find parameters")
		return err
	}
	t.Parameters = []ConfigParameter{}
	for _, rawMessage := range rawParameters {
		var m map[string]interface{}
		err = json.Unmarshal(*rawMessage, &m)
		if err != nil {
			log.Println("could not unmarshal parameters")
			return err
		}
		if _, ok := m["use"]; ok { // derivedParam
			p := DerivedParameter{}
			err = json.Unmarshal(*rawMessage, &p)
			if err != nil {
				log.Println("could not unmarshal determined parameter")
				return err
			}
			t.Parameters = append(t.Parameters, p)
		} else if _, ok := m["name"]; ok { // basicParam
			p := BasicParameter{}
			err = json.Unmarshal(*rawMessage, &p)
			if err != nil {
				log.Println("could not unmarshal determined parameter")
				return err
			}
			t.Parameters = append(t.Parameters, p)
		}

	}

	// Deserialize Steps by type
	var rawSteps []*json.RawMessage
	err = json.Unmarshal(*objMap["steps"], &rawSteps)
	if err != nil {
		log.Println("could not find steps")
		return err
	}
	t.Steps = []Step{}
	var m map[string]interface{}
	for _, rawMessage := range rawSteps {
		err = json.Unmarshal(*rawMessage, &m)
		if err != nil {
			log.Println("could not unmarshal step")
			return err
		}

		s, err := DetermineStepFromInterface(m)
		if err != nil {
			log.Println(err)
		} else {
			err := json.Unmarshal(*rawMessage, s)
			if err != nil {
				log.Println(err)
			} else {
				t.Steps = append(t.Steps, s)
			}
		}

	}

	return nil
}

func (t *Task) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var taskMap map[string]interface{}
	err := unmarshal(&taskMap)
	if err != nil {
		log.Printf("unable to unmarshal task")
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

	t.Parameters = []ConfigParameter{}
	switch x := taskMap["parameters"].(type) {
	case []interface{}:
		for _, p := range x {
			switch y := p.(type) {
			case map[interface{}]interface{}:
				if _, ok := y["use"]; ok { // derivedParam
					var dP DerivedParameter
					dP.UnmarshalYamlInterface(y)
					t.Parameters = append(t.Parameters, dP)
				} else if _, ok := y["name"]; ok { // basicParam
					var bP BasicParameter
					bP.UnmarshalYamlInterface(y)
					t.Parameters = append(t.Parameters, bP)
				}
				break
			}
		}
		break
	}

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
					log.Println(err)
				} else {
					err = s.UnmarshalYamlInterface(y)
					if err != nil {
						log.Println(err)
					} else {
						t.Steps = append(t.Steps, s)
					}
				}
				break
			}
		}
		break
	}

	// log.Printf("Unmarshalled Task: %+v", t)
	return nil
}
