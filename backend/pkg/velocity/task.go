package velocity

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

type Task struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Git         TaskGit           `json:"git" yaml:"git"`
	Docker      TaskDocker        `json:"docker" yaml:"docker"`
	Parameters  []ParameterConfig `json:"parameters" yaml:"parameters"`
	Steps       []Step            `json:"steps" yaml:"steps"`

	ValidationErrors   []string `json:"validationErrors" yaml:"-"`
	ValidationWarnings []string `json:"validationWarnings" yaml:"-"`

	ProjectRoot        string               `json:"-" yaml:"-"`
	RunID              string               `json:"-" yaml:"-"`
	ResolvedParameters map[string]Parameter `json:"-" yaml:"-"`
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
		// GetLogger().Error("could not unmarshal task", zap.Error(err))
		return err
	}

	t.ValidationErrors = []string{}
	t.ValidationWarnings = []string{}

	switch x := taskMap["description"].(type) {
	case string:
		t.Description = x
		break
	default:
		t.ValidationWarnings = append(t.ValidationWarnings, "missing 'description'")
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
					err := d.UnmarshalYamlInterface(z)
					if err != nil {
						t.ValidationErrors = append(t.ValidationErrors, err.Error())
						// return err
					}
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
					// GetLogger().Error("could not determine step from interface", zap.Error(err))
					t.ValidationErrors = append(t.ValidationErrors, err.Error())
				} else {
					s.SetProjectRoot(t.ProjectRoot)
					err = s.UnmarshalYamlInterface(y)
					if err != nil {
						// GetLogger().Error("could not unmarshal yaml step", zap.Error(err))
						t.ValidationErrors = append(t.ValidationErrors, err.Error())
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

func findProjectRoot(cwd string, attempted []string) (string, error) {
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.IsDir() && f.Name() == ".git" {
			GetLogger().Debug("found project root", zap.String("dir", cwd))
			return cwd, nil
		}
	}

	if filepath.Dir(cwd) == cwd {
		return "", fmt.Errorf("could not find project root. Tried: %v", append(attempted, cwd))
	}

	return findProjectRoot(filepath.Dir(cwd), append(attempted, cwd))
}

func findTasksDirectory(projectRoot string) (string, error) {
	// fmt.Printf("checking %s for tasks directory.\n", cwd)
	tasksDir := "tasks"

	// check for tasks setting in velocity.yml
	repoConfigPath := filepath.Join(projectRoot, ".velocity.yml")
	if f, err := os.Stat(repoConfigPath); !os.IsNotExist(err) {
		if !f.IsDir() {
			var repoConfig RepositoryConfig
			repoYaml, _ := ioutil.ReadFile(repoConfigPath)
			err = yaml.Unmarshal(repoYaml, &repoConfig)
			if err == nil {
				if repoConfig.Project.TasksPath != "" {
					tasksDir = repoConfig.Project.TasksPath
				}
			}
		}
	}

	tasksPath := filepath.Join(projectRoot, tasksDir)
	if f, err := os.Stat(tasksPath); !os.IsNotExist(err) {
		if f.IsDir() {
			return tasksPath, nil
		}
	}

	return "", fmt.Errorf("could not find tasks in: %s", filepath.Join(projectRoot, tasksDir))
}

func GetTasksFromCurrentDir() ([]Task, error) {
	tasks := []Task{}

	cwd, err := os.Getwd()
	if err != nil {
		return tasks, err
	}

	projectRoot, err := findProjectRoot(cwd, []string{})
	if err != nil {
		return tasks, err
	}

	tasksDir, err := findTasksDirectory(projectRoot)
	if err != nil {
		return tasks, err
	}

	GetLogger().Debug("looking for tasks in", zap.String("dir", tasksDir))
	err = filepath.Walk(tasksDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml")) {
			// fmt.Printf("-> reading %s\n", path)
			taskYml, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var t Task
			relativePath, err := filepath.Rel(tasksDir, path)
			if err != nil {
				return err
			}
			t.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
			t.ProjectRoot = projectRoot
			err = yaml.Unmarshal(taskYml, &t)
			if err != nil {
				return err
			}
			tasks = append(tasks, t)
		}
		return nil
	})

	return tasks, err
}
