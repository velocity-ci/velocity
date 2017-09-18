package task

import (
	yaml "gopkg.in/yaml.v2"
)

type yamlTask struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Parameters  []Parameter              `yaml:"parameters"`
	Steps       []map[string]interface{} `yaml:"steps"`
}

func ResolveTaskFromYAML(y string, additionalParams []Parameter) Task {
	yTask := yamlTask{
		Name:        "",
		Description: "",
		Parameters:  []Parameter{},
	}
	err := yaml.Unmarshal([]byte(y), &yTask)
	if err != nil {
		panic(err)
	}

	task := Task{
		Name:        yTask.Name,
		Description: yTask.Description,
		Parameters:  yTask.Parameters,
		Steps:       []Step{},
	}

	for _, yStep := range yTask.Steps {
		mStep, err := yaml.Marshal(yStep)
		if err != nil {
			panic(err)
		}
		s := ResolveStepFromYAML(string(mStep[:]))
		err = s.Validate(append(task.Parameters, additionalParams...))
		if err != nil {
			panic(err)
		}
		task.Steps = append(task.Steps, s)
	}
	return task
}
