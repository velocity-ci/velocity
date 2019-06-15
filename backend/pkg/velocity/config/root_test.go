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
  blueprintsPath: ./blueprints
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

stages: 
- name: test
  blueprints: 
  - test_web
  - test_api
- name: release
  blueprints: 
  - release_web
  - release_api
- name: deploy
  blueprints: 
  - deploy_web
  - deploy_api	
`

	var repositoryConfig config.Root
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	logo := string("image.png")
	expectedRepositoryConfig := config.Root{
		Project: &config.RootProject{
			Logo:           &logo,
			BlueprintsPath: "./blueprints",
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
		Stages: []*config.RootStage{
			{
				Name:       "test",
				Blueprints: []string{"test_web", "test_api"},
			},
			{
				Name:       "release",
				Blueprints: []string{"release_web", "release_api"},
			},
			{
				Name:       "deploy",
				Blueprints: []string{"deploy_web", "deploy_api"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, repositoryConfig)
}

func TestRepositoryConfigUnmarshalLogoNil(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
  blueprintsPath: ./blueprints
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

stages: 
- name: test
  blueprints: 
  - test_web
  - test_api
- name: release
  blueprints: 
  - release_web
  - release_api
- name: deploy
  blueprints: 
  - deploy_web
  - deploy_api	
`

	var repositoryConfig config.Root
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	expectedRepositoryConfig := &config.Root{
		Project: &config.RootProject{
			Logo:           nil,
			BlueprintsPath: "./blueprints",
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
		Stages: []*config.RootStage{
			{
				Name:       "test",
				Blueprints: []string{"test_web", "test_api"},
			},
			{
				Name:       "release",
				Blueprints: []string{"release_web", "release_api"},
			},
			{
				Name:       "deploy",
				Blueprints: []string{"deploy_web", "deploy_api"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, &repositoryConfig)
}
