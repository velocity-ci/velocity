package step

import (
	"fmt"

	"github.com/velocity-ci/velocity/master/velocity/domain"
)

type DockerCompose struct {
	domain.BaseStep
	ComposeFile string `json:"composeFile" yaml:"compose_file"`
	Contents    dockerComposeYaml
}

func (dC *DockerCompose) SetEmitter(e func(string)) {
	dC.Emit = e
}

func (dC DockerCompose) GetType() string {
	return "compose"
}

func (dC DockerCompose) GetDescription() string {
	return dC.Description
}

func (dC DockerCompose) GetDetails() string {
	return fmt.Sprintf("composeFile: %s", dC.ComposeFile)
}

func (dC *DockerCompose) Validate(params []domain.Parameter) error {
	return nil
}

func (dC *DockerCompose) SetParams(params []domain.Parameter) error {
	return nil
}

func (dC *DockerCompose) Execute() error {

	return nil
}

type dockerComposeYaml struct {
	Services map[string]dockerComposeService `json:"services" yaml:"services"`
}

type dockerComposeService struct {
	Build       dockerComposeServiceBuild `json:"build" yaml:"build"`
	WorkingDir  string                    `json:"workingDir" yaml:"working_dir"`
	Command     string                    `json:"command" yaml:"command"`
	Links       []string                  `json:"links" yaml:"links"`
	Environment map[string]string         `json:"environment" yaml:"environment"`
	Volumes     []string                  `json:"volumes" yaml:"volumes"`
	Expose      []string                  `json:"expose" yaml:"expose"`
}

type dockerComposeServiceBuild struct {
	Context    string `json:"context" yaml:"context"`
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`
}