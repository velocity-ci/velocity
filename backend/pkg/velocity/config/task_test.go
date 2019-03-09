package config

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestTaskConfigUnmarshal(t *testing.T) {
	taskConfigYaml := `
---
description: "Hello Velocity"

parameters:
  - use: https://velocityci.io/parameter-test
    arguments:
      name: /velocityci/foo
    exports:
      value: bar
  - name: your_name
    secret: true

docker:
  registries:
    - use: https://velocityci.io/registry-test
      arguments:
        username: registry_user
        password: registry_password
`
	taskConfig := newTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := newTask()
	expectedTaskConfig.Description = "Hello Velocity"
	expectedTaskConfig.Parameters = []Parameter{
		&ParameterDerived{
			BaseParameter: BaseParameter{Type: "derived"},
			Use:           "https://velocityci.io/parameter-test",
			Arguments: map[string]string{
				"name": "/velocityci/foo",
			},
			Exports: map[string]string{
				"value": "bar",
			},
		},
		&ParameterBasic{
			BaseParameter: BaseParameter{Type: "basic"},
			Name:          "your_name",
			Secret:        true,
		},
	}
	expectedTaskConfig.Docker = TaskDocker{
		Registries: []TaskDockerRegistry{
			TaskDockerRegistry{
				Address: "",
				Use:     "https://velocityci.io/registry-test",
				Arguments: map[string]string{
					"username": "registry_user",
					"password": "registry_password",
				},
			},
		},
	}

	assert.Equal(t, expectedTaskConfig, taskConfig)
}
