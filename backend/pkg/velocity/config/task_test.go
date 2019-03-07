package config_test

import (
	"testing"

	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"

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
	taskConfig := config.NewTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := config.NewTask()
	expectedTaskConfig.Description = "Hello Velocity"
	expectedTaskConfig.Parameters = []config.Parameter{
		&config.ParameterDerived{
			BaseParameter: config.BaseParameter{Type: "derived"},
			Use:           "https://velocityci.io/parameter-test",
			Arguments: map[string]string{
				"name": "/velocityci/foo",
			},
			Exports: map[string]string{
				"value": "bar",
			},
		},
		&config.ParameterBasic{
			BaseParameter: config.BaseParameter{Type: "basic"},
			Name:          "your_name",
			Secret:        true,
		},
	}
	expectedTaskConfig.Docker = config.TaskDocker{
		Registries: []config.TaskDockerRegistry{
			config.TaskDockerRegistry{
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
