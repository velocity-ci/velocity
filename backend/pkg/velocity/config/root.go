package config

import "encoding/json"

type Root struct {
	Project *RootProject `json:"project"`
	Git     *RootGit     `json:"git"`

	Parameters []Parameter   `json:"parameters"`
	Plugins    []*RootPlugin `json:"plugins"`
	Stages     []*RootStage  `json:"stages"`
}

type RootProject struct {
	Logo      *string `json:"logo"`
	TasksPath string  `json:"tasksPath"`
}

type RootGit struct {
	// Depth     int  `json:"depth"`
	Submodule bool `json:"submodule"`
}

type RootPlugin struct {
	Use       string            `json:"use"`
	Arguments map[string]string `json:"arguments"`
	Events    []string          `json:"events"`
}

type RootStage struct {
	Name  string   `json:"name"`
	Tasks []string `json:"tasks"`
}

func (r *Root) UnmarshalJSON(b []byte) error {
	// We don't return any errors from this function so we can show more helpful parse errors
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	// Deserialize Project
	if _, ok := objMap["project"]; ok {
		err = json.Unmarshal(*objMap["project"], &r.Project)
		if err != nil {
			return err
		}
	}

	// Deserialize Git
	if _, ok := objMap["git"]; ok {
		err = json.Unmarshal(*objMap["git"], &r.Git)
		if err != nil {
			return err
		}
	}

	// Deserialize Parameters
	if val, _ := objMap["parameters"]; val != nil {
		var rawParameters []*json.RawMessage
		err = json.Unmarshal(*val, &rawParameters)
		if err != nil {
			return err
		}
		if err == nil {
			for _, rawMessage := range rawParameters {
				param, err := unmarshalParameter(*rawMessage)
				if err != nil {
					return err
				}
				r.Parameters = append(r.Parameters, param)
			}
		}
	}

	// Deserialize Git
	if _, ok := objMap["plugins"]; ok {
		err = json.Unmarshal(*objMap["plugins"], &r.Plugins)
		if err != nil {
			return err
		}
	}

	// Deserialize Git
	if _, ok := objMap["stages"]; ok {
		err = json.Unmarshal(*objMap["stages"], &r.Stages)
		if err != nil {
			return err
		}
	}
	return nil
}
