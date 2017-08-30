package step

import (
	"log"
	"testing"
)

func TestResolveStepFromYAML(t *testing.T) {
	runStep := `type: run
description: Initialise Terraform
image: hashicorp/terraform
command: terraform init
environment:
  TFVAR_ENVIRONMENT: ${e}
`

	step := ResolveStepFromYAML(runStep)

	log.Println(step)

	if step.GetDescription() != "Initialise Terraform" {
		t.Fail()
	}

}
