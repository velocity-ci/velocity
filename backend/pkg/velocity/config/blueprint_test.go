package config

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

func TestBlueprintConfigUnmarshal(t *testing.T) {
	blueprintConfigYaml := `
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
	blueprintConfig := newBlueprint()

	err := yaml.Unmarshal([]byte(blueprintConfigYaml), &blueprintConfig)
	assert.Nil(t, err)

	expectedBlueprintConfig := newBlueprint()
	expectedBlueprintConfig.Description = "Hello Velocity"
	// expectedBlueprintConfig.Parameters = []Parameter{
	// 	&ParameterDerived{
	// 		BaseParameter: BaseParameter{Type: "derived"},
	// 		Use:           "https://velocityci.io/parameter-test",
	// 		Arguments: map[string]string{
	// 			"name": "/velocityci/foo",
	// 		},
	// 		Exports: map[string]string{
	// 			"value": "bar",
	// 		},
	// 	},
	// 	&ParameterBasic{
	// 		BaseParameter: BaseParameter{Type: "basic"},
	// 		Name:          "your_name",
	// 		Secret:        true,
	// 	},
	// }
	// expectedBlueprintConfig.Docker = BlueprintDocker{
	// 	Registries: []BlueprintDockerRegistry{
	// 		{
	// 			Address: "",
	// 			Use:     "https://velocityci.io/registry-test",
	// 			Arguments: map[string]string{
	// 				"username": "registry_user",
	// 				"password": "registry_password",
	// 			},
	// 		},
	// 	},
	// }

	assert.Equal(t, expectedBlueprintConfig, blueprintConfig)
}
