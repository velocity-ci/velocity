package velocity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
	"gopkg.in/yaml.v2"
)

func TestTaskConfigUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: "Hello Velocity"
name: hello-velocity

`
	var taskConfig velocity.Task

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.Task{
		Name:        "hello-velocity",
		Description: "Hello Velocity",
		Steps:       []velocity.Step{},
		Docker: velocity.TaskDocker{
			Registries: []velocity.DockerRegistry{},
		},
		Parameters: []velocity.ParameterConfig{},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}
