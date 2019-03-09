package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
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

func newTask() *Task {
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

func findTasksDirectory(root *Root) (string, error) {
	tasksDir := "tasks"
	if root.Project.TasksPath != "" {
		tasksDir = root.Project.TasksPath
	}

	tasksPath := filepath.Join(root.Path, tasksDir)
	if f, err := os.Stat(tasksPath); !os.IsNotExist(err) {
		if f.IsDir() {
			return tasksPath, nil
		}
	}

	return "", fmt.Errorf("could not find tasks in: %s", tasksPath)
}

func GetTasksFromRoot(root *Root) ([]*Task, error) {
	tasks := []*Task{}

	tasksPath, err := findTasksDirectory(root)
	if err != nil {
		return tasks, err
	}

	logging.GetLogger().Debug("looking for tasks in", zap.String("dir", tasksPath))
	err = filepath.Walk(tasksPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml")) {
			// fmt.Printf("-> reading %s\n", path)
			taskYml, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			t := newTask()
			relativePath, err := filepath.Rel(tasksPath, path)
			if err != nil {
				return err
			}
			t.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
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
