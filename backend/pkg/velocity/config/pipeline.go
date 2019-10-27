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

type Stage struct {
	Name       string   `json:"name"`
	Blueprints []string `json:"blueprints"`
}

func (s *Stage) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	if _, ok := objMap["name"]; ok {
		err = json.Unmarshal(*objMap["name"], &s.Name)
		if err != nil {
			return err
		}
	}

	if _, ok := objMap["blueprints"]; ok {
		err = json.Unmarshal(*objMap["blueprints"], &s.Blueprints)
		if err != nil {
			return err
		}
	}

	return nil
}

type pipeline struct {
	Name        string `json:"name"`
	Description string `json:"description"`

	Stages []*Stage `json:"stages"`

	ParseErrors      []string `json:"parseErrors"`
	ValidationErrors []string `json:"validationErrors"`
}

func newPipeline() *pipeline {
	return &pipeline{
		Name:             "",
		Description:      "",
		Stages:           []*Stage{},
		ParseErrors:      []string{},
		ValidationErrors: []string{},
	}
}

func handlePipelineUnmarshalError(t *pipeline, err error) *pipeline {
	if err != nil {
		t.ParseErrors = append(t.ParseErrors, err.Error())
	}

	return t
}

func (t *pipeline) UnmarshalJSON(b []byte) error {
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

	// Default pipeline
	if t.Name == "default" {
		t.Description = "The default pipeline"
	}

	// Deserialize Stages
	if val, _ := objMap["stages"]; val != nil {
		var rawStages []*json.RawMessage
		err = json.Unmarshal(*val, &rawStages)
		t = handlePipelineUnmarshalError(t, err)
		if err == nil {
			for i, rawMessage := range rawStages {
				s := &Stage{}
				err = json.Unmarshal(*rawMessage, s)
				t = handlePipelineUnmarshalError(t, err)
				if err == nil {
					if s.Name == "" {
						s.Name = fmt.Sprintf("stage %d", i)
					}
					t.Stages = append(t.Stages, s)
				}
			}
		}
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

func GetPipelinesFromRoot(root *Root) ([]*v1.Pipeline, error) {
	pipelines := []*v1.Pipeline{}

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
			cP := newPipeline()
			relativePath, err := filepath.Rel(pipelinesPath, path)
			if err != nil {
				return err
			}
			cP.Name = strings.TrimSuffix(relativePath, filepath.Ext(relativePath))
			err = yaml.Unmarshal(pipelineYml, &cP)
			if err != nil {
				return err
			}
			pipelines = append(pipelines, parsePipeline(cP, root.Repository.CurrentCommitInfo))
		}
		return nil
	})

	return pipelines, err
}

func parsePipeline(cP *pipeline, cm *git.RawCommit) *v1.Pipeline {
	return &v1.Pipeline{
		Id:          fmt.Sprintf("%s+%s", cm.SHA, cP.Name),
		ProjectId:   "",
		Name:        cP.Name,
		CommitId:    cm.SHA,
		Description: cP.Description,
	}
}
