package velocity

import (
	yaml "gopkg.in/yaml.v2"
)

func ResolveTaskFromYAML(y string, additionalParams map[string]Parameter) Task {
	var t Task
	err := yaml.Unmarshal([]byte(y), &t)
	if err != nil {
		panic(err)
	}
	// allParams := map[string]Parameter{}
	// for k, v := range task.Parameters {
	// 	allParams[k] = v
	// }
	// for k, v := range additionalParams {
	// 	allParams[k] = v
	// }

	// for _, yStep := range t.Steps {
	// 	mStep, err := yaml.Marshal(yStep)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	s := ResolveStepFromYAML(string(mStep[:]))
	// 	if s != nil {
	// 		// err = s.Validate(allParams)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		s.SetParams(additionalParams)
	// 		task.Steps = append(t.Steps, s)
	// 	} else {
	// 		log.Printf("failed to resolve step: %s", yStep)
	// 	}
	// }
	return t
}
