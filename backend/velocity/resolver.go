package velocity

import (
	"log"

	yaml "gopkg.in/yaml.v2"
)

type yamlTask struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Parameters  []ConfigParameter        `yaml:"parameters"`
	Steps       []map[string]interface{} `yaml:"steps"`
}

func ResolveTaskFromYAML(y string, additionalParams map[string]Parameter) Task {
	yTask := yamlTask{
		Name:        "",
		Description: "",
		Parameters:  []ConfigParameter{},
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

	// allParams := map[string]Parameter{}
	// for k, v := range task.Parameters {
	// 	allParams[k] = v
	// }
	// for k, v := range additionalParams {
	// 	allParams[k] = v
	// }

	for _, yStep := range yTask.Steps {
		mStep, err := yaml.Marshal(yStep)
		if err != nil {
			panic(err)
		}
		s := ResolveStepFromYAML(string(mStep[:]))
		if s != nil {
			// err = s.Validate(allParams)
			if err != nil {
				panic(err)
			}
			s.SetParams(additionalParams)
			task.Steps = append(task.Steps, s)
		} else {
			log.Printf("failed to resolve step: %s", yStep)
		}
	}
	return task
}
