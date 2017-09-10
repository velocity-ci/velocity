package task

import (
	"log"
	"testing"
)

func TestResolveStepFromYAML(t *testing.T) {
	taskYaml := `name: Deploy
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
`

	task := ResolveTaskFromYAML(taskYaml)

	log.Println(task)

	if task.Name != "Deploy" ||
		task.Description != "Deploys application" ||
		len(task.Parameters) != 1 ||
		task.Parameters[0].Name != "e" ||
		task.Parameters[0].Value != "testing" ||
		len(task.Parameters[0].OtherOptions) != 1 ||
		task.Parameters[0].OtherOptions[0] != "production" ||
		len(task.Steps) != 2 ||
		task.Steps[0].GetType() != "run" ||
		task.Steps[0].GetDescription() != "Initialise Terraform" ||
		task.Steps[1].GetType() != "run" ||
		task.Steps[1].GetDescription() != "Plan Terraform" {
		t.Fail()
	}

}
