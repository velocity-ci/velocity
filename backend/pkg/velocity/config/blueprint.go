package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/git"
	v1 "github.com/velocity-ci/velocity/backend/pkg/velocity/genproto/v1"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

type blueprint struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Docker      blueprintDocker `json:"docker"`
	Parameters  []parameter     `json:"parameters"`
	Steps       []step          `json:"steps"`

	ParseErrors      []string `json:"parseErrors"`
	ValidationErrors []string `json:"validationErrors"`
}

func newBlueprint() *blueprint {
	return &blueprint{
		Name:        "",
		Description: "",
		Docker: blueprintDocker{
			Registries: []blueprintDockerRegistry{},
		},
		Parameters:       []parameter{},
		Steps:            []step{},
		ParseErrors:      []string{},
		ValidationErrors: []string{},
	}
}

func handleBlueprintUnmarshalError(t *blueprint, err error) *blueprint {
	if err != nil {
		t.ParseErrors = append(t.ParseErrors, err.Error())
	}

	return t
}

// UnmarshalJSON provides custom JSON decoding
func (t *blueprint) UnmarshalJSON(b []byte) error {
	// We don't return any errors from this function so we can show more helpful parse errors
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		t = handleBlueprintUnmarshalError(t, err)
	}

	// Deserialize Description
	if _, ok := objMap["description"]; ok {
		err = json.Unmarshal(*objMap["description"], &t.Description)
		t = handleBlueprintUnmarshalError(t, err)
	}

	// Deserialize Parameters
	if val, _ := objMap["parameters"]; val != nil {
		var rawParameters []*json.RawMessage
		err = json.Unmarshal(*val, &rawParameters)
		t = handleBlueprintUnmarshalError(t, err)
		if err == nil {
			for _, rawMessage := range rawParameters {
				param, err := unmarshalParameter(*rawMessage)
				t = handleBlueprintUnmarshalError(t, err)
				if param != nil {
					t.Parameters = append(t.Parameters, param)
				}
			}
		}
	}

	// Deserialize Docker
	if _, ok := objMap["docker"]; ok {
		err = json.Unmarshal(*objMap["docker"], &t.Docker)
		t = handleBlueprintUnmarshalError(t, err)
	}

	// Deserialize Steps by type
	if val, _ := objMap["steps"]; val != nil {
		var rawSteps []*json.RawMessage
		err = json.Unmarshal(*val, &rawSteps)
		t = handleBlueprintUnmarshalError(t, err)
		if err == nil {
			for _, rawMessage := range rawSteps {
				s, err := unmarshalStep(*rawMessage)
				t = handleBlueprintUnmarshalError(t, err)
				if err == nil {
					err = json.Unmarshal(*rawMessage, s)
					t = handleBlueprintUnmarshalError(t, err)
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
	blueprintsDir := filepath.Join(root.Project.ConfigPath, "blueprints")

	blueprintsPath := filepath.Join(root.Path, blueprintsDir)
	if f, err := os.Stat(blueprintsPath); !os.IsNotExist(err) {
		if f.IsDir() {
			return blueprintsPath, nil
		}
	}

	return "", fmt.Errorf("could not find blueprints in: %s", blueprintsPath)
}

// GetBlueprintsFromRoot returns the blueprints found from a given root directory
func GetBlueprintsFromRoot(root *Root) ([]*v1.Blueprint, error) {
	blueprints := []*v1.Blueprint{}

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
			cB := newBlueprint()
			relativePath, err := filepath.Rel(blueprintsPath, path)
			if err != nil {
				return err
			}
			err = yaml.Unmarshal(blueprintYml, &cB)
			if err != nil {
				return err
			}
			cB.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
			b, err := parseBlueprint(cB, root.Repository.CurrentCommitInfo)
			if err != nil {
				return err
			}
			blueprints = append(blueprints, b)
		}
		return nil
	})

	return blueprints, err
}

func parseBlueprint(cB *blueprint, cm *git.RawCommit) (*v1.Blueprint, error) {
	steps := []*v1.Step{}
	for _, cS := range cB.Steps {
		s, err := parseStep(cS)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}
	return &v1.Blueprint{
		Id:          fmt.Sprintf("%s+%s", cm.SHA, cB.Name),
		ProjectId:   "",
		Name:        cB.Name,
		CommitId:    cm.SHA,
		Description: cB.Description,
		Steps:       steps,
	}, nil
}
