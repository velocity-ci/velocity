package velocity

import (
	"log"
	"testing"
)

func TestResolveStepFromYAML(t *testing.T) {

	stepSpec := []string{
		`
type: run
description: Initialise Terraform
image: hashicorp/terraform
working_dir: ./api
command: ["terraform", "init"]
environment:
  TFVAR_ENVIRONMENT: ${e}
`,
	}

	for _, step := range stepSpec {
		s := ResolveStepFromYAML(step)

		if s.GetType() != "run" {
			log.Println(s)
			t.Fail()
		}
	}

}
