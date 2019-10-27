package config

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestDockerBuildUnmarshal(t *testing.T) {
	blueprintConfigYaml := `
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
	blueprintConfig := newBlueprint()
	err := yaml.Unmarshal([]byte(blueprintConfigYaml), &blueprintConfig)
	assert.Nil(t, err)

	expectedBlueprintConfig := newBlueprint()
	expectedBlueprintConfig.Description = "Builds a docker image"
	// expectedBlueprintConfig.Steps = []Step{
	// 	&StepDockerBuild{
	// 		BaseStep: BaseStep{
	// 			Type:        "build",
	// 			Description: "Docker build",
	// 		},
	// 		Dockerfile: "test.Dockerfile",
	// 		Context:    "./test",
	// 		Tags: []string{
	// 			"test/a:333",
	// 			"test/b:344",
	// 		},
	// 	},
	// }

	assert.EqualValues(t, expectedBlueprintConfig, blueprintConfig)
}

func TestDockerComposeUnmarshal(t *testing.T) {
	blueprintConfigYaml := `
---
description: Runs integration tests
steps:
  - type: compose
    description: Docker compose
    composeFile: test.docker-compose.yml
`
	blueprintConfig := newBlueprint()
	err := yaml.Unmarshal([]byte(blueprintConfigYaml), &blueprintConfig)
	assert.Nil(t, err)

	expectedBlueprintConfig := newBlueprint()
	expectedBlueprintConfig.Description = "Runs integration tests"
	// expectedBlueprintConfig.Steps = []Step{
	// 	&StepDockerCompose{
	// 		BaseStep: BaseStep{
	// 			Type:        "compose",
	// 			Description: "Docker compose",
	// 		},
	// 		ComposeFile: "test.docker-compose.yml",
	// 	},
	// }

	assert.EqualValues(t, expectedBlueprintConfig, blueprintConfig)
}

func TestDockerPushUnmarshal(t *testing.T) {
	blueprintConfigYaml := `
---
description: Pushes a docker container
steps:
  - type: push
    description: Docker push
    tags:
      - test/a:333
      - test/b:344
`
	blueprintConfig := newBlueprint()
	err := yaml.Unmarshal([]byte(blueprintConfigYaml), &blueprintConfig)
	assert.Nil(t, err)

	expectedBlueprintConfig := newBlueprint()
	expectedBlueprintConfig.Description = "Pushes a docker container"
	// expectedBlueprintConfig.Steps = []Step{
	// 	&StepDockerPush{
	// 		BaseStep: BaseStep{
	// 			Type:        "push",
	// 			Description: "Docker push",
	// 		},
	// 		Tags: []string{
	// 			"test/a:333",
	// 			"test/b:344",
	// 		},
	// 	},
	// }

	assert.Equal(t, expectedBlueprintConfig, blueprintConfig)
}

func TestDockerRunUnmarshal(t *testing.T) {
	blueprintConfigYaml := `
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
	blueprintConfig := newBlueprint()

	err := yaml.Unmarshal([]byte(blueprintConfigYaml), blueprintConfig)
	assert.Nil(t, err)

	expectedBlueprintConfig := newBlueprint()
	expectedBlueprintConfig.Description = "Runs a docker container"
	// expectedBlueprintConfig.Steps = []Step{
	// 	&StepDockerRun{
	// 		BaseStep: BaseStep{
	// 			Type:        "run",
	// 			Description: "Hello Docker",
	// 		},
	// 		Image:   "hello-world:latest",
	// 		Command: []string{"sleep", "10"},
	// 		Environment: map[string]string{
	// 			"HELLO": "WORLD",
	// 		},
	// 		WorkingDir:     "/app",
	// 		MountPoint:     "/app",
	// 		IgnoreExitCode: false,
	// 	},
	// 	&StepDockerRun{
	// 		BaseStep: BaseStep{
	// 			Type:        "run",
	// 			Description: "Hello Array Environment",
	// 		},
	// 		Image:   "hello-world:latest",
	// 		Command: []string{},
	// 		Environment: map[string]string{
	// 			"HELLO": "WORLD",
	// 		},
	// 	},
	// }

	assert.Equal(t, expectedBlueprintConfig, blueprintConfig)
}
