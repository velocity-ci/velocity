package task

import (
	"github.com/velocity-ci/velocity/master/velocity/domain"
	"github.com/velocity-ci/velocity/master/velocity/domain/step"
	yaml "gopkg.in/yaml.v2"
)

type yamlTask struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Parameters  []domain.Parameter       `yaml:"parameters"`
	Steps       []map[string]interface{} `yaml:"steps"`
}

func ResolveTaskFromYAML(y string, additionalParams []domain.Parameter) domain.Task {
	yTask := yamlTask{}
	err := yaml.Unmarshal([]byte(y), &yTask)
	if err != nil {
		panic(err)
	}

	task := domain.Task{
		Name:        yTask.Name,
		Description: yTask.Description,
		Parameters:  yTask.Parameters,
	}

	for _, yStep := range yTask.Steps {
		mStep, err := yaml.Marshal(yStep)
		if err != nil {
			panic(err)
		}
		s := step.ResolveStepFromYAML(string(mStep[:]))
		err = s.Validate(append(task.Parameters, additionalParams...))
		if err != nil {
			panic(err)
		}
		task.Steps = append(task.Steps, s)
	}
	return task
}
