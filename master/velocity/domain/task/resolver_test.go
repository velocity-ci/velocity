package task

import (
	"log"
	"reflect"
	"testing"
)

type taskTestSpec struct {
	val               string
	derivedParameters []Parameter
	expected          Task
}

func TestResolveTaskFromYAML(t *testing.T) {

	taskSpecs := []taskTestSpec{
		taskTestSpec{
			val: `
name: Deploy
description: Deploys application

parameters:
  - name: e
    default: testing
    other_options:
      - production

steps:
  - type: run
    description: Initialise Terraform
    image: hashicorp/terraform
    command: ["terraform", "init"]
    environment:
      TFVAR_ENVIRONMENT: ${e}
  - type: run
    description: Plan Terraform
    image: hashicorp/terraform
    command: ["terraform", "plan"]
    environment:
      TFVAR_ENVIRONMENT: ${e}
      TFVAR_IMAGE_TAG: ${GIT_SHA}
`,
			derivedParameters: []Parameter{
				Parameter{
					Name:  "GIT_SHA",
					Value: "test_SHA",
				},
			},
			expected: Task{
				Name:        "Deploy",
				Description: "Deploys application",
				Parameters: []Parameter{
					Parameter{
						Name:         "e",
						Value:        "testing",
						OtherOptions: []string{"production"},
						Secret:       false,
					},
				},
				Steps: []Step{
					&DockerRun{
						BaseStep: BaseStep{
							Type:        "run",
							Description: "Initialise Terraform",
						},
						Image:          "hashicorp/terraform",
						Command:        []string{"terraform", "init"},
						Environment:    map[string]string{"TFVAR_ENVIRONMENT": "${e}"},
						WorkingDir:     "",
						MountPoint:     "",
						IgnoreExitCode: false,
					},
					&DockerRun{
						BaseStep: BaseStep{
							Type:        "run",
							Description: "Plan Terraform",
						},
						Image:          "hashicorp/terraform",
						Command:        []string{"terraform", "plan"},
						Environment:    map[string]string{"TFVAR_ENVIRONMENT": "${e}", "TFVAR_IMAGE_TAG": "test_SHA"},
						WorkingDir:     "",
						MountPoint:     "",
						IgnoreExitCode: false,
					},
				},
			},
		},
	}

	for _, taskSpec := range taskSpecs {
		ta := ResolveTaskFromYAML(taskSpec.val, taskSpec.derivedParameters)
		if !reflect.DeepEqual(ta, taskSpec.expected) {
			log.Println(taskSpec.expected)
			log.Println("!=")
			log.Println(ta)
			t.Fail()
		}
	}

}
