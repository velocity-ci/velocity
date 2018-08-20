package velocity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"gopkg.in/yaml.v2"
)

func TestDockerRunUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
name: docker-run

steps:
  - type: run
    description: Hello Docker
    image: hello-world:latest
`
	var taskConfig velocity.Task

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.Task{
		Name:        "docker-run",
		Description: "",
		Steps: []velocity.Step{
			&velocity.DockerRun{
				BaseStep: velocity.BaseStep{
					Type:          "run",
					Description:   "Hello Docker",
					OutputStreams: []string{"run"},
					Params:        map[string]velocity.Parameter{},
				},
				Image:       "hello-world:latest",
				Command:     []string{},
				Environment: map[string]string{},
			},
		},
		Docker: velocity.TaskDocker{
			Registries: []velocity.DockerRegistry{},
		},
		Parameters: []velocity.ParameterConfig{},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}
