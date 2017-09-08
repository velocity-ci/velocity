package step

import (
	"github.com/velocity-ci/velocity/master/velocity/domain"
	yaml "gopkg.in/yaml.v2"
)

func ResolveStepFromYAML(y string) domain.Step {
	bStep := BaseStep{}
	err := yaml.Unmarshal([]byte(y), &bStep)
	if err != nil {
		panic(err)
	}

	switch bStep.Type {
	case "run":
		return resolveRunStep(y)
	default:
		return nil
	}
}

func resolveRunStep(y string) domain.Step {
	step := &DockerRun{}
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return step
}
