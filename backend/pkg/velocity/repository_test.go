package velocity_test

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/velocity-ci/velocity/backend/pkg/velocity"
)

func TestRepositoryConfigUnmarshal(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
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

	var repositoryConfig velocity.RepositoryConfig
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), &repositoryConfig)

	assert.Nil(t, err)
	logo := string("image.png")
	expectedRepositoryConfig := velocity.RepositoryConfig{
		Project: &velocity.ProjectConfig{
			Logo:      &logo,
			TasksPath: "./tasks",
		},
		Git: &velocity.GitConfig{
			Submodule: false,
		},
		Parameters: []velocity.ParameterConfig{
			&velocity.DerivedParameter{
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
		Plugins: []*velocity.PluginConfig{
			&velocity.PluginConfig{
				Use: "param-slack-bin-uri",
				Arguments: map[string]string{
					"CHANNEL":     "ci",
					"SLACK_TOKEN": "${slack}",
				},
				Events: []string{"BUILD_START", "BUILD_FAIL", "BUILD_SUCCESS"},
			},
		},
		Stages: []*velocity.StageConfig{
			&velocity.StageConfig{
				Name:  "test",
				Tasks: []string{"test_web", "test_api"},
			},
			&velocity.StageConfig{
				Name:  "release",
				Tasks: []string{"release_web", "release_api"},
			},
			&velocity.StageConfig{
				Name:  "deploy",
				Tasks: []string{"deploy_web", "deploy_api"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, repositoryConfig)
}

func TestRepositoryConfigUnmarshalLogoNil(t *testing.T) {
	repositoryConfigYaml := `
---
project: 
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

	var repositoryConfig = velocity.NewRepositoryConfig()
	err := yaml.Unmarshal([]byte(repositoryConfigYaml), repositoryConfig)

	assert.Nil(t, err)
	expectedRepositoryConfig := &velocity.RepositoryConfig{
		Project: &velocity.ProjectConfig{
			Logo:      nil,
			TasksPath: "./tasks",
		},
		Git: &velocity.GitConfig{
			Submodule: false,
		},
		Parameters: []velocity.ParameterConfig{
			&velocity.DerivedParameter{
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
		Plugins: []*velocity.PluginConfig{
			&velocity.PluginConfig{
				Use: "param-slack-bin-uri",
				Arguments: map[string]string{
					"CHANNEL":     "ci",
					"SLACK_TOKEN": "${slack}",
				},
				Events: []string{"BUILD_START", "BUILD_FAIL", "BUILD_SUCCESS"},
			},
		},
		Stages: []*velocity.StageConfig{
			&velocity.StageConfig{
				Name:  "test",
				Tasks: []string{"test_web", "test_api"},
			},
			&velocity.StageConfig{
				Name:  "release",
				Tasks: []string{"release_web", "release_api"},
			},
			&velocity.StageConfig{
				Name:  "deploy",
				Tasks: []string{"deploy_web", "deploy_api"},
			},
		},
	}
	assert.Equal(t, expectedRepositoryConfig, repositoryConfig)
}
