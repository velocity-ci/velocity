package task_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
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
	taskConfig := velocity.NewTask()

	err := yaml.Unmarshal([]byte(taskConfigYaml), &taskConfig)
	assert.Nil(t, err)

	expectedTaskConfig := velocity.NewTask()
	expectedTaskConfig.Description = "Hello Velocity"
	expectedTaskConfig.Parameters = []velocity.ParameterConfig{
		&velocity.DerivedParameter{
			Type: "derived",
			Use:  "https://velocityci.io/parameter-test",
			Arguments: map[string]string{
				"name": "/velocityci/foo",
			},
			Exports: map[string]string{
				"value": "bar",
			},
		},
		&velocity.BasicParameter{
			Type:   "basic",
			Name:   "your_name",
			Secret: true,
		},
	}
	expectedTaskConfig.Docker = velocity.TaskDocker{
		Registries: []velocity.DockerRegistry{
			velocity.DockerRegistry{
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
