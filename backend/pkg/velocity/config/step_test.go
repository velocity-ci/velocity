package config_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
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
	taskConfig := config.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := config.NewTask()
	expectedTaskConfig.Description = "Builds a docker image"
	expectedTaskConfig.Steps = []config.Step{
		&config.StepDockerBuild{
			BaseStep: config.BaseStep{
				Type:        "build",
				Description: "Docker build",
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

func TestDockerComposeUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: Runs integration tests
steps:
  - type: compose 
    description: Docker compose
    composeFile: test.docker-compose.yml
`
	taskConfig := config.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := config.NewTask()
	expectedTaskConfig.Description = "Runs integration tests"
	expectedTaskConfig.Steps = []config.Step{
		&config.StepDockerCompose{
			BaseStep: config.BaseStep{
				Type:        "compose",
				Description: "Docker compose",
			},
			ComposeFile: "test.docker-compose.yml",
		},
	}

	assert.EqualValues(t, expectedTaskConfig, taskConfig)
}

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
	taskConfig := config.NewTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := config.NewTask()
	expectedTaskConfig.Description = "Pushes a docker container"
	expectedTaskConfig.Steps = []config.Step{
		&config.StepDockerPush{
			BaseStep: config.BaseStep{
				Type:        "push",
				Description: "Docker push",
			},
			Tags: []string{
				"test/a:333",
				"test/b:344",
			},
		},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}

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
	taskConfig := config.NewTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := config.NewTask()
	expectedTaskConfig.Description = "Runs a docker container"
	expectedTaskConfig.Steps = []config.Step{
		&config.StepDockerRun{
			BaseStep: config.BaseStep{
				Type:        "run",
				Description: "Hello Docker",
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
		&config.StepDockerRun{
			BaseStep: config.BaseStep{
				Type:        "run",
				Description: "Hello Array Environment",
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
