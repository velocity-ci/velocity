package task

import (
	"log"
	"testing"
)

func TestResolveStepFromYAML(t *testing.T) {
	runStep := `type: run
description: Initialise Terraform
image: hashicorp/terraform
working_dir: ./api
command: ["terraform", "init"]
environment:
  TFVAR_ENVIRONMENT: ${e}
`

	step := ResolveStepFromYAML(runStep)

	log.Println(step)

	if step.GetType() != "run" {
		t.Fail()
	}
}
