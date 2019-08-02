package config_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity/config"
)

func TestRootUnmarshal(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
  logo: image.png
  configPath: .velocity
git: 
  depth: 10

parameters: 
- use: param-s3-bin-uri
  arguments: 
    uri: "s3://mybucket/project_secrets"
  exports: 
   slack_token: slack
  secret: true

plugins: 
- use: plugin-slack-bin-uri
  arguments: 
    CHANNEL: ci
    SLACK_TOKEN: "${slack}"
  events: 
  - BUILD_START
  - BUILD_FAIL
  - BUILD_SUCCESS
`

	var repositoryConfig config.Root
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	logo := string("image.png")
	expectedRepositoryConfig := config.Root{
		Project: &config.RootProject{
			Logo:       &logo,
			ConfigPath: ".velocity",
		},
		Git: &config.RootGit{
			Submodule: false,
		},
		Parameters: []config.Parameter{
			&config.ParameterDerived{
				BaseParameter: config.BaseParameter{Type: "derived"},
				Use:           "param-s3-bin-uri",
				Arguments: map[string]string{
					"uri": "s3://mybucket/project_secrets",
				},
				Exports: map[string]string{
					"slack_token": "slack",
				},
				Secret: true,
			},
		},
		Plugins: []*config.RootPlugin{
			{
				Use: "plugin-slack-bin-uri",
				Arguments: map[string]string{
					"CHANNEL":     "ci",
					"SLACK_TOKEN": "${slack}",
				},
				Events: []string{"BUILD_START", "BUILD_FAIL", "BUILD_SUCCESS"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, repositoryConfig)
}

func TestRepositoryConfigUnmarshalLogoNil(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
  configPath: .velocity
git: 
  depth: 10

parameters: 
- use: param-s3-bin-uri
  arguments: 
    uri: "s3://mybucket/project_secrets"
  exports: 
   slack_token: slack
  secret: true

plugins: 
- use: param-slack-bin-uri
  arguments: 
    CHANNEL: ci
    SLACK_TOKEN: "${slack}"
  events: 
  - BUILD_START
  - BUILD_FAIL
  - BUILD_SUCCESS
`

	var repositoryConfig config.Root
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	expectedRepositoryConfig := &config.Root{
		Project: &config.RootProject{
			Logo:       nil,
			ConfigPath: ".velocity",
		},
		Git: &config.RootGit{
			Submodule: false,
		},
		Parameters: []config.Parameter{
			&config.ParameterDerived{
				BaseParameter: config.BaseParameter{Type: "derived"},
				Use:           "param-s3-bin-uri",
				Arguments: map[string]string{
					"uri": "s3://mybucket/project_secrets",
				},
				Exports: map[string]string{
					"slack_token": "slack",
				},
				Secret: true,
			},
		},
		Plugins: []*config.RootPlugin{
			{
				Use: "param-slack-bin-uri",
				Arguments: map[string]string{
					"CHANNEL":     "ci",
					"SLACK_TOKEN": "${slack}",
				},
				Events: []string{"BUILD_START", "BUILD_FAIL", "BUILD_SUCCESS"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, &repositoryConfig)
}
