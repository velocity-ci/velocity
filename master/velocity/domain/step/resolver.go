package step

import (
	"github.com/velocity-ci/velocity/master/velocity/domain"
	yaml "gopkg.in/yaml.v2"
)

func ResolveStepFromYAML(y string) domain.Step {
	bStep := domain.BaseStep{}
	err := yaml.Unmarshal([]byte(y), &bStep)
	if err != nil {
		panic(err)
	}

	switch bStep.Type {
	case "run":
		return resolveRunStep(y)
	case "build":
		return resolveBuildStep(y)
	case "plugin":
		return resolvePluginStep(y)
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

func resolveBuildStep(y string) domain.Step {
	step := &DockerBuild{}
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return step
}

func resolvePluginStep(y string) domain.Step {
	step := &Plugin{}
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return step
}
