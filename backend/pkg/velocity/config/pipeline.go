package config

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/logging"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Pipeline struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	ParseErrors      []string `json:"parseErrors"`
	ValidationErrors []string `json:"validationErrors"`
}

func newPipeline() *Pipeline {
	return &Pipeline{
		Name:        "",
		Description: "",
	}
}

func handlePipelineUnmarshalError(t *Pipeline, err error) *Pipeline {
	if err != nil {
		t.ParseErrors = append(t.ParseErrors, err.Error())
	}

	return t
}

func (t *Pipeline) UnmarshalJSON(b []byte) error {
	// We don't return any errors from this function so we can show more helpful parse errors
	var objMap map[string]*json.RawMessage
	// We'll store the error (if any) so we can return it if necessary
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		t = handlePipelineUnmarshalError(t, err)
	}

	// Deserialize Description
	if _, ok := objMap["description"]; ok {
		err = json.Unmarshal(*objMap["description"], &t.Description)
		t = handlePipelineUnmarshalError(t, err)
	}

	// Default Pipeline
	if t.Name == "default" {
		t.Description = "The default pipeline"
	}

	return nil
}

func findPipelinesDirectory(root *Root) (string, error) {
	pipelinesDir := filepath.Join(root.Project.ConfigPath, "pipelines")

	pipelinesPath := filepath.Join(root.Path, pipelinesDir)
	if f, err := os.Stat(pipelinesPath); !os.IsNotExist(err) {
		if f.IsDir() {
			return pipelinesPath, nil
		}
	}

	return "", fmt.Errorf("could not find pipelines in: %s", pipelinesPath)
}

func GetPipelinesFromRoot(root *Root) ([]*Pipeline, error) {
	pipelines := []*Pipeline{}

	pipelinesPath, err := findPipelinesDirectory(root)
	if err != nil {
		return pipelines, err
	}

	logging.GetLogger().Debug("looking for pipelines in", zap.String("dir", pipelinesPath))
	err = filepath.Walk(pipelinesPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml")) {
			// fmt.Printf("-> reading %s\n", path)
			pipelineYml, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			t := newPipeline()
			relativePath, err := filepath.Rel(pipelinesPath, path)
			if err != nil {
				return err
			}
			t.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
			err = yaml.Unmarshal(pipelineYml, &t)
			if err != nil {
				return err
			}
			pipelines = append(pipelines, t)
		}
		return nil
	})

	return pipelines, err
}
