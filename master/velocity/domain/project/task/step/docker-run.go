package step

import (
	"fmt"
	"regexp"

	"github.com/VJftw/velocity/master/velocity/domain"
)

type DockerRun struct {
	BaseStep
	Description string            `json:"description" yaml:"description"`
	Image       string            `json:"image" yaml:"image"`
	Command     string            `json:"command" yaml:"command"`
	Environment map[string]string `json:"environment" yaml:"environment"`
}

func (dB DockerRun) GetType() string {
	return "run"
}

func (dB DockerRun) GetDescription() string {
	return dB.Description
}

func (dB DockerRun) GetDetails() string {
	return fmt.Sprintf("image: %s command: %s", dB.Image, dB.Command)
}

func (dB DockerRun) Execute() error {
	return nil
}

func (dR DockerRun) Validate(params []domain.Parameter) error {
	re := regexp.MustCompile("\\$\\{(.+)\\}")

	requiredParams := re.FindAllStringSubmatch(dR.Image, -1)
	if !isAllInParams(requiredParams, params) {
		return fmt.Errorf("Parameter %v missing", requiredParams)
	}
	requiredParams = re.FindAllStringSubmatch(dR.Command, -1)
	if !isAllInParams(requiredParams, params) {
		return fmt.Errorf("Parameter %v missing", requiredParams)
	}
	for key, val := range dR.Environment {
		requiredParams = re.FindAllStringSubmatch(key, -1)
		if !isAllInParams(requiredParams, params) {
			return fmt.Errorf("Parameter %v missing", requiredParams)
		}
		requiredParams = re.FindAllStringSubmatch(val, -1)
		if !isAllInParams(requiredParams, params) {
			return fmt.Errorf("Parameter %v missing", requiredParams)
		}
	}
	return nil
}

func isAllInParams(matches [][]string, params []domain.Parameter) bool {
	for _, match := range matches {
		found := false
		for _, param := range params {
			if param.Name == match[1] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
