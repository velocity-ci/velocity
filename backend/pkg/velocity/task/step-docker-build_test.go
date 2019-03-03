package task_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func TestDockerBuildUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: Builds a docker image
steps:
  - type: build
    description: Docker build
    dockerfile: test.Dockerfile
    context: ./test
    tags:
      - test/a:333
      - test/b:344
`
	taskConfig := velocity.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.NewTask()
	expectedTaskConfig.Description = "Builds a docker image"
	expectedTaskConfig.Steps = []velocity.Step{
		&velocity.DockerBuild{
			BaseStep: velocity.BaseStep{
				Type:          "build",
				Description:   "Docker build",
				OutputStreams: []string{"build"},
				Status:        "waiting",
			},
			Dockerfile: "test.Dockerfile",
			Context:    "./test",
			Tags: []string{
				"test/a:333",
				"test/b:344",
			},
		},
	}

	assert.EqualValues(t, expectedTaskConfig, taskConfig)
}
