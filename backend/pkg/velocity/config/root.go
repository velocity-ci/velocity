package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
)

type Root struct {
	Path    string       `json:"-"`
	Project *RootProject `json:"project"`
	Git     *RootGit     `json:"git"`

	Parameters []Parameter   `json:"parameters"`
	Plugins    []*RootPlugin `json:"plugins"`
	Stages     []*RootStage  `json:"stages"`
}

type RootProject struct {
	Logo           *string `json:"logo"`
	BlueprintsPath string  `json:"blueprintsPath"`
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
	Name       string   `json:"name"`
	Blueprints []string `json:"blueprints"`
}

func newRoot() *Root {
	return &Root{
		Project: &RootProject{},
		Git: &RootGit{
			Submodule: true,
		},
		Parameters: []Parameter{},
		Plugins:    []*RootPlugin{},
		Stages:     []*RootStage{},
	}
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

func GetRootConfig() (*Root, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	projectRoot, err := findProjectRoot(cwd, []string{})
	if err != nil {
		return nil, err
	}

	rootConfig := newRoot()
	rootConfigPath := filepath.Join(projectRoot, ".velocity.yml")
	if f, err := os.Stat(rootConfigPath); !os.IsNotExist(err) {
		if !f.IsDir() {
			repoYaml, _ := ioutil.ReadFile(rootConfigPath)
			err := yaml.Unmarshal(repoYaml, rootConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	rootConfig.Path = projectRoot
	return rootConfig, nil
}

func findProjectRoot(cwd string, attempted []string) (string, error) {
	files, err := ioutil.ReadDir(cwd)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		if f.IsDir() && f.Name() == ".git" {
			logging.GetLogger().Debug("found project root", zap.String("dir", cwd))
			return cwd, nil
		}
	}

	if filepath.Dir(cwd) == cwd {
		return "", fmt.Errorf("could not find project root. Tried: %v", append(attempted, cwd))
	}

	return findProjectRoot(filepath.Dir(cwd), append(attempted, cwd))
}
