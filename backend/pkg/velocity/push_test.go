package velocity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	yaml "gopkg.in/yaml.v2"
)

func TestDockerPushUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
name: docker-push

steps:
  - type: push
    description: Docker push
    tags:
      - test/a:333
      - test/b:344
`
	var taskConfig velocity.Task

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.Task{
		Name:        "docker-push",
		Description: "",
		Steps: []velocity.Step{
			&velocity.DockerPush{
				BaseStep: velocity.BaseStep{
					Type:          "push",
					Description:   "Docker push",
					OutputStreams: []string{"push"},
					Params:        map[string]velocity.Parameter{},
				},
				Tags: []string{
					"test/a:333",
					"test/b:344",
				},
			},
		},
		Docker: velocity.TaskDocker{
			Registries: []velocity.DockerRegistry{},
		},
		Parameters: []velocity.ParameterConfig{},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}