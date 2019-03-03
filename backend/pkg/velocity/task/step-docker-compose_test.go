package task_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func TestDockerComposeUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: Runs integration tests
steps:
  - type: compose 
    description: Docker compose
    composeFile: test.docker-compose.yml
`
	taskConfig := velocity.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.NewTask()
	expectedTaskConfig.Description = "Runs integration tests"
	expectedTaskConfig.Steps = []velocity.Step{
		&velocity.DockerCompose{
			BaseStep: velocity.BaseStep{
				Type:          "compose",
				Description:   "Docker compose",
				OutputStreams: []string{},
				Status:        "waiting",
			},
			ComposeFile: "test.docker-compose.yml",
		},
	}

	assert.EqualValues(t, expectedTaskConfig, taskConfig)
}
