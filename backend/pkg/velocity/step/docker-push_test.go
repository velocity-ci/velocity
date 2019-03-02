package step_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func TestDockerPushUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: Pushes a docker container
steps:
  - type: push
    description: Docker push
    tags:
      - test/a:333
      - test/b:344
`
	taskConfig := velocity.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.NewTask()
	expectedTaskConfig.Description = "Pushes a docker container"
	expectedTaskConfig.Steps = []velocity.Step{
		&velocity.DockerPush{
			BaseStep: velocity.BaseStep{
				Type:          "push",
				Description:   "Docker push",
				OutputStreams: []string{"push"},
				Status:        "waiting",
			},
			Tags: []string{
				"test/a:333",
				"test/b:344",
			},
		},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}
