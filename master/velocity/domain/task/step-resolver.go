package task

import (
	yaml "gopkg.in/yaml.v2"
)

func ResolveStepFromYAML(y string) Step {
	bStep := BaseStep{}
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

func resolveRunStep(y string) Step {
	step := NewDockerRun()
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return &step
}

func resolveBuildStep(y string) Step {
	step := NewDockerBuild()
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return &step
}

func resolvePluginStep(y string) Step {
	step := &Plugin{}
	err := yaml.Unmarshal([]byte(y), &step)
	if err != nil {
		panic(err)
	}
	return step
}