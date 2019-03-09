package config

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
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
	taskConfig := newTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := newTask()
	expectedTaskConfig.Description = "Builds a docker image"
	expectedTaskConfig.Steps = []Step{
		&StepDockerBuild{
			BaseStep: BaseStep{
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
	taskConfig := newTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := newTask()
	expectedTaskConfig.Description = "Runs integration tests"
	expectedTaskConfig.Steps = []Step{
		&StepDockerCompose{
			BaseStep: BaseStep{
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
	taskConfig := newTask()
	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := newTask()
	expectedTaskConfig.Description = "Pushes a docker container"
	expectedTaskConfig.Steps = []Step{
		&StepDockerPush{
			BaseStep: BaseStep{
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
	taskConfig := newTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := newTask()
	expectedTaskConfig.Description = "Runs a docker container"
	expectedTaskConfig.Steps = []Step{
		&StepDockerRun{
			BaseStep: BaseStep{
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
		&StepDockerRun{
			BaseStep: BaseStep{
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
