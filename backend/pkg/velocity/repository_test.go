package velocity_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestRepositoryConfigUnmarshal(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
  name: Velocity
  logo: image.png
  tasksPath: ./tasks
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
  tasks: 
  - test_web
  - test_api
- name: release
  tasks: 
  - release_web
  - release_api
- name: deploy
  tasks: 
  - deploy_web
  - deploy_api	
`

	var repositoryConfig RepositoryConfig
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	expectedRepositoryConfig := RepositoryConfig{
		Project: ProjectConfig{
			Name:      "Velocity",
			Logo:      "image.png",
			TasksPath: "./tasks",
		},
		Git: GitConfig{
			Depth: 10,
		},
		Parameters: []ParameterConfig{
			DerivedParameter{
				Type: "derived",
				Use:  "param-s3-bin-uri",
				Arguments: map[string]string{
					"uri": "s3://mybucket/project_secrets",
				},
				Exports: map[string]string{
					"slack_token": "slack",
				},
				Secret: true,
			},
		},
		Plugins: []PluginConfig{
			PluginConfig{
				Use: "param-slack-bin-uri",
				Arguments: map[string]string{
					"CHANNEL":     "ci",
					"SLACK_TOKEN": "${slack}",
				},
				Events: []string{"BUILD_START", "BUILD_FAIL", "BUILD_SUCCESS"},
			},
		},
		Stages: []StageConfig{
			StageConfig{
				Name:  "test",
				Tasks: []string{"test_web", "test_api"},
			},
			StageConfig{
				Name:  "release",
				Tasks: []string{"release_web", "release_api"},
			},
			StageConfig{
				Name:  "deploy",
				Tasks: []string{"deploy_web", "deploy_api"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, repositoryConfig)
}
