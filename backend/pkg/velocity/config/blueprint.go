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

type Blueprint struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Docker      BlueprintDocker `json:"docker"`
	Parameters  []Parameter     `json:"parameters"`
	Steps       []Step          `json:"steps"`

	ParseErrors      []string `json:"parseErrors"`
	ValidationErrors []string `json:"validationErrors"`
}

func newBlueprint() *Blueprint {
	return &Blueprint{
		Name:        "",
		Description: "",
		Docker: BlueprintDocker{
			Registries: []BlueprintDockerRegistry{},
		},
		Parameters:       []Parameter{},
		Steps:            []Step{},
		ParseErrors:      []string{},
		ValidationErrors: []string{},
	}
}

func handleUnmarshalError(t *Blueprint, err error) *Blueprint {
	if err != nil {
		t.ParseErrors = append(t.ParseErrors, err.Error())
	}

	return t
}

func (t *Blueprint) UnmarshalJSON(b []byte) error {
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

func findBlueprintsDirectory(root *Root) (string, error) {
	blueprintsDir := filepath.Join(".velocityci", "blueprints")
	if root.Project.BlueprintsPath != "" {
		blueprintsDir = root.Project.BlueprintsPath
	}

	blueprintsPath := filepath.Join(root.Path, blueprintsDir)
	if f, err := os.Stat(blueprintsPath); !os.IsNotExist(err) {
		if f.IsDir() {
			return blueprintsPath, nil
		}
	}

	return "", fmt.Errorf("could not find blueprints in: %s", blueprintsPath)
}

func GetBlueprintsFromRoot(root *Root) ([]*Blueprint, error) {
	blueprints := []*Blueprint{}

	blueprintsPath, err := findBlueprintsDirectory(root)
	if err != nil {
		return blueprints, err
	}

	logging.GetLogger().Debug("looking for blueprints in", zap.String("dir", blueprintsPath))
	err = filepath.Walk(blueprintsPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml")) {
			// fmt.Printf("-> reading %s\n", path)
			blueprintYml, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			t := newBlueprint()
			relativePath, err := filepath.Rel(blueprintsPath, path)
			if err != nil {
				return err
			}
			t.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
			err = yaml.Unmarshal(blueprintYml, &t)
			if err != nil {
				return err
			}
			blueprints = append(blueprints, t)
		}
		return nil
	})

	return blueprints, err
}
