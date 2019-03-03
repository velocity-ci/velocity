package task_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func TestDockerRunUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: Runs a docker container
steps:
  - type: run
    description: Hello Docker
    image: hello-world:latest
    command: sleep 10
    environment:
      HELLO: WORLD
    workingDir: /app
    mountPoint: /app
    ignoreExitCode: false
  - type: run
    description: Hello Array Environment
    image: hello-world:latest
    environment:
     - HELLO=WORLD
`
	taskConfig := velocity.NewTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.NewTask()
	expectedTaskConfig.Description = "Runs a docker container"
	expectedTaskConfig.Steps = []velocity.Step{
		&velocity.DockerRun{
			BaseStep: velocity.BaseStep{
				Type:          "run",
				Description:   "Hello Docker",
				OutputStreams: []string{"run"},
				Status:        "waiting",
			},
			Image:   "hello-world:latest",
			Command: []string{"sleep", "10"},
			Environment: map[string]string{
				"HELLO": "WORLD",
			},
			WorkingDir:     "/app",
			MountPoint:     "/app",
			IgnoreExitCode: false,
		},
		&velocity.DockerRun{
			BaseStep: velocity.BaseStep{
				Type:          "run",
				Description:   "Hello Array Environment",
				OutputStreams: []string{"run"},
				Status:        "waiting",
			},
			Image:   "hello-world:latest",
			Command: []string{},
			Environment: map[string]string{
				"HELLO": "WORLD",
			},
		},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}
